# Connectors

Connectors are the components which connect two objects to facilitate the flow of events between them. Example: A connector could connect a cron trigger to docker scaler. This will ensure that the desired number of docker containers are running in a machine periodically.

Connectors can connect a trigger to a trigger (and form a chain of triggers) that ultimately is connected to a scaler. Connectors are also a good place to do data transformation of data from a trigger to a format of data that a scaler can understand.

| type   | status    | description                                                    |
| ------ | --------- | -------------------------------------------------------------- |
| direct | Available | Directly pass events between Trigger-Scaler or Trigger-Trigger |

Propose a new connector [here](https://github.com/scriptnull/waymond/issues/new).
