# jira2zulip-pm-adapter
When a user is tagged in a comment to a task in JIRA - a notification is sent to the PM to the user who was mentioned from "JIRA-Bot" with a link to the task in which the user (~user) was mentioned and the body of the message.

When there are many tasks and a large development is underway, this speeds up the communication time through ZULIP for joint production of tasks.

The server is raised to wait for webhooks on GO on port 5000.
