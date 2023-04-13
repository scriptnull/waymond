---
sidebar_position: 2
---

# Concepts

### Triggers

Triggers are the components which could trigger an event, which could ultimately yield to autoscaling. The decision of when and how to auto-scale will flow from here.

### Scalers

Scalers are the components which perform autoscaling operation. They access the target APIs (like docker or AWS API) to achieve a desired state in those systems.

### Connectors

Connectors are the components which connect two objects to facilitate the flow of events between them. Example: A connector could connect a cron trigger to docker scaler. This will ensure that the desired number of docker containers are running in a machine periodically.

Connectors can connect a trigger to a trigger (and form a chain of triggers) that ultimately is connected to a scaler. Connectors are also a good place to do data transformation of data from a trigger to a format of data that a scaler can understand.

### Event Bus
All the components `triggers`, `scalers`, and `connectors` are internally connected via a simple event bus (don't be scared it is just a Go channel and some helper functions :smile:). This event-based architecture will help any of the above mentioned components to capture and act on events in a seamless way.

### Architecture

![architecture](https://user-images.githubusercontent.com/4211715/222922530-fda823c7-1a72-4156-99ac-3d249e4e8e47.png)