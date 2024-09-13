# Redmine Activity Notifier

Redmine Activity Notifier is a Go program that fetches activity updates from a Redmine instance and sends notifications to a Slack channel. It allows you to stay informed about the latest changes in your Redmine projects directly in your Slack workspace.

## Features

- Fetches activity updates from a Redmine Atom feed at specified intervals
- Sends notifications to a designated Slack channel
- Supports basic authentication for the Redmine API
- Configurable through a YAML configuration file

## Prerequisites

- A Redmine instance with an accessible Atom feed
- A Slack workspace with an incoming webhook set up

## Installation

```
$ go build
```

## Configuration

- `interval`: The interval at which to fetch updates from the Redmine Atom feed (e.g., 15m for 15 minutes)
- `slack_url`: The URL of your Slack incoming webhook
- `atom_url`: The URL of your Redmine Atom feed
- `basic_auth` (optional): If your Redmine instance requires basic authentication, provide the username and password

## Usage

```
$ ./redmine-activity-notifier
```

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
