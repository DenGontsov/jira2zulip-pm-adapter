package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "regexp"
    "strings"
)

// Конфигурация Zulip
const (
    ZULIP_API_URL   = "https://chat.youorg-zulip.tld/api/v1/messages" // Базовый URL для ZULIP API
    ZULIP_BOT_EMAIL = "jira-bot@chat.youorg-zulip.tld"                // Имя пользователя для JIRA-bot в Вашем ZULIP
    ZULIP_API_KEY   = "<PASTE_YOUR_ZULIP_API_KEY_HERE>"               // API-ключ для ZULIP бота
)

// Конфигурация JIRA
const (
    JIRA_API_URL  = "https://jira.youorg.tld/rest/api/2/issue/" // Базовый URL для JIRA API
    JIRA_USERNAME = "tech-jira-user"                            // Имя пользователя для JIRA
    JIRA_PASSWORD = "<PASTE_YOUR_JIRA_PASS_HERE>"               // Пароль для JIRA
)

// Функция для отправки сообщения в Zulip
func sendToZulip(toUsername, content string) bool {
    data := fmt.Sprintf("type=private&to=%s@ad.youorg.tld&content=%s", toUsername, content)
    log.Printf("Отправляем данные в Zulip: %s", data)

    req, err := http.NewRequest("POST", ZULIP_API_URL, strings.NewReader(data))
    if err != nil {
        log.Printf("Ошибка при создании запроса к Zulip: %v\n", err)
        return false
    }

    req.SetBasicAuth(ZULIP_BOT_EMAIL, ZULIP_API_KEY)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Printf("Ошибка при отправке запроса в Zulip: %v\n", err)
        return false
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Ошибка при чтении ответа от Zulip: %v\n", err)
        return false
    }

    log.Printf("Ответ от Zulip: %s", body)

    if resp.StatusCode != http.StatusOK {
        log.Printf("Ошибка отправки сообщения в Zulip: %s\n", resp.Status)
        return false
    }

    return true
}

// Функция для извлечения ID задачи из URL комментария
func extractIssueID(issueURL string) (string, error) {
    re := regexp.MustCompile(`/issue/(\d+)`)
    matches := re.FindStringSubmatch(issueURL)
    if len(matches) < 2 {
        return "", fmt.Errorf("не удалось извлечь ID задачи из URL")
    }
    return matches[1], nil
}

// Функция для извлечения ключа задачи из JIRA API
func fetchIssueKey(issueID string) (string, error) {
    url := JIRA_API_URL + issueID

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return "", fmt.Errorf("ошибка при создании запроса к JIRA: %v", err)
    }

    req.SetBasicAuth(JIRA_USERNAME, JIRA_PASSWORD)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("ошибка при отправке запроса к JIRA: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("не удалось получить данные задачи, статус: %s", resp.Status)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("ошибка чтения ответа от JIRA: %v", err)
    }

    var issueData map[string]interface{}
    if err := json.Unmarshal(body, &issueData); err != nil {
        return "", fmt.Errorf("ошибка разбора JSON-ответа от JIRA: %v", err)
    }

    issueKey, ok := issueData["key"].(string)
    if !ok {
        return "", fmt.Errorf("ключ задачи отсутствует в ответе JIRA")
    }

    return issueKey, nil
}

// Функция для извлечения имени пользователя из комментария
func extractMentionedUser(commentBody string) string {
    re := regexp.MustCompile(`~([a-zA-Z0-9_-]+)`)
    matches := re.FindStringSubmatch(commentBody)
    if len(matches) > 1 {
        return matches[1]
    }
    return ""
}

// Обработка вебхука
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
        return
    }

    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, `{"error":"Ошибка чтения тела запроса"}`, http.StatusBadRequest)
        return
    }

    var data map[string]interface{}
    if err := json.Unmarshal(body, &data); err != nil {
        http.Error(w, `{"error":"Получены некорректные данные"}`, http.StatusBadRequest)
        return
    }

    commentData, ok := data["comment"].(map[string]interface{})
    if !ok {
        http.Error(w, `{"error":"Отсутствует или неверный формат поля 'comment'"}`, http.StatusBadRequest)
        return
    }

    commentURL, ok := commentData["self"].(string)
    if !ok {
        http.Error(w, `{"error":"Отсутствует или неверный формат поля 'self' в 'comment'"}`, http.StatusBadRequest)
        return
    }

    issueID, err := extractIssueID(commentURL)
    if err != nil {
        http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
        return
    }

    issueKey, err := fetchIssueKey(issueID)
    if err != nil {
        http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
        return
    }

    commentBody, ok := commentData["body"].(string)
    if !ok {
        http.Error(w, `{"error":"Отсутствует или неверный формат поля 'body' в 'comment'"}`, http.StatusBadRequest)
        return
    }

    authorData, ok := commentData["author"].(map[string]interface{})
    if !ok {
        http.Error(w, `{"error":"Отсутствует или неверный формат поля 'author' в 'comment'"}`, http.StatusBadRequest)
        return
    }

    authorName, ok := authorData["name"].(string)
    if !ok {
        http.Error(w, `{"error":"Отсутствует или неверный формат поля 'name' в 'author'"}`, http.StatusBadRequest)
        return
    }

    mentionedUser := extractMentionedUser(commentBody)
    if mentionedUser == "" {
        http.Error(w, `{"error":"Не найдено упомянутое имя пользователя в комментарии"}`, http.StatusBadRequest)
        return
    }

    content := fmt.Sprintf(
        "**%s упомянул вас в комментарие к задаче** [%s](https://jira.youorg.tld/browse/%s):\n\n%s",
        authorName,
        issueKey,
        issueKey,
        commentBody,
    )

    if sendToZulip(mentionedUser, content) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"Сообщение отправлено в Zulip"}`))
    } else {
        http.Error(w, `{"error":"Не удалось отправить сообщение в Zulip"}`, http.StatusInternalServerError)
    }
}

func main() {
    http.HandleFunc("/api", handleWebhook)

    port := "5000"
    log.Printf("Автор: den@gontsov.net \n")
    log.Printf("Дополнительная информация и использование: https://github.com/DenGontsov/jira2zulip-pm-adapter \n")
    log.Printf("Сервер запущен на порту %s...\n", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatalf("Ошибка запуска сервера: %v\n", err)
    }
}
