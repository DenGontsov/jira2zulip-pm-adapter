# jira2zulip-pm-adapter
When a user is tagged in a comment to a task in JIRA - a notification is sent to the PM to the user who was mentioned from "JIRA-Bot" with a link to the task in which the user (~user) was mentioned and the body of the message.

When there are many tasks and a large development is underway, this speeds up the communication time through Zulip for joint production of tasks.

The server is raised to wait for webhooks on GO on port 5000.

---

Когда пользователь отмечен в комментарии к задаче в JIRA - в ЛС пользователю в Zulip, которого упомянули, отправляется уведомление из "JIRA-Bot" со ссылкой на задачу, в которой был упомянут пользователь (~user) и текстом сообщения.

Когда задач много и идет активная разработка, это ускоряет время общения через Zulip для совместного производства.

Сервер поднимается для ожидания вебхуков на GO на порту 5000.

## BUILD
```
git clone https://github.com/DenGontsov/jira2zulip-pm-adapter
cd jira2zulip-pm-adapter/
go build jira2zulip-pm-adapter.go
```

## TECH SPECs
```
1) ad.youorg.tld - replace with yours in the code. In Zulip, the user is identified by username@ad.youorg.tld, if you use integration with AD.
2) Before the build, you need to create a bot in Zulip and create a technical user in JIRA (This is necessary to get the task key by its ID).
3) Create a webhook in JIRA that looks at the IP or bound domain where the instance of the currently collected instance will be launched.

```
