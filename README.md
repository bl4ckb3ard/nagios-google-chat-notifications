# Nagios Google Chat Notifications

A fast, lightweight, and feature-rich Golang binary for sending Nagios Core alerts directly to Google Chat spaces using the official Google Chat API and Cards V2 formatting. 

## Features
- **Rich Cards V2 Layout:** Clean, structured visual data that is easy to scan.
- **Dynamic Visuals:** Automatically changes the emoji (🟢, ⚠️, 🚨, ❓, ✅, 💤, 🔀) based on the exact state, notification type, and problem status.
- **Direct Nagios Links:** Generates URL-encoded links to jump straight to the specific host or service status CGI page.
- **Actionable Buttons:** Conditionally adds an "Acknowledge" button to active problems, routing users straight to the `cmd.cgi` submission page.
- **Audit Trails:** Catches and displays Nagios Acknowledgement and Scheduled Downtime notifications, showing exactly who took ownership/scheduled it and their comment.
- **Targeted Escalations:** Uses a command-line flag (`--mention-all`) to drop a loud `<users/all>` ping for high-priority Nagios escalations.

---

## Installation

### 1. Clone and Build
Make sure you have Go installed on your system.

```bash
git clone https://github.com/bl4ckb3ard/nagios-google-chat-notifications.git
cd nagios-google-chat-notifications
go mod tidy
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o nagios-google-chat-notifications main.go
```

### 2. Move to Nagios Plugin Directory
Copy the compiled binary to your Nagios `libexec` directory and ensure proper permissions.

```bash
sudo cp nagios-google-chat-notifications /usr/local/nagios/libexec/
sudo chown nagios:nagios /usr/local/nagios/libexec/nagios-google-chat-notifications
sudo chmod +x /usr/local/nagios/libexec/nagios-google-chat-notifications
```

---

## Usage & CLI Flags

The binary accepts the following command-line flags. Most of these map directly to standard Nagios macros.

| Flag                 | Required | Nagios Macro                       | Description                                                                    |
| :------------------- | :------: | :--------------------------------- | :----------------------------------------------------------------------------- |
| `--webhook`          | Yes      | `$_CONTACTGOOGLE_CHAT_WEBHOOK$`    | The Google Chat Webhook URL.                                                   |
| `--type`             | Yes      | N/A                                | Must be either `host` or `service`.                                            |
| `--nagios-url`       | No       | N/A                                | Base URL for your Nagios Web UI (e.g., `https://your-nagios-base-url/nagios`). |
| `--notification-type`| No       | `$NOTIFICATIONTYPE$`               | Type of alert (PROBLEM, RECOVERY, ACKNOWLEDGEMENT, etc.)                       |
| `--hostname`         | No       | `$HOSTNAME$`                       | Name of the host.                                                              |
| `--hostaddress`      | No       | `$HOSTADDRESS$`                    | IP Address of the host.                                                        |
| `--state`            | No       | `$HOSTSTATE$` / `$SERVICESTATE$`   | State of the alert (UP, DOWN, OK, WARNING, CRITICAL).                          |
| `--service-desc`     | No       | `$SERVICEDESC$`                    | Description of the service (Service alerts only).                              |
| `--output`           | No       | `$HOSTOUTPUT$` / `$SERVICEOUTPUT$` | The plugin output text.                                                        |
| `--datetime`         | No       | `$LONGDATETIME$`                   | Timestamp of the alert.                                                        |
| `--author`           | No       | `$NOTIFICATIONAUTHOR$`             | User who acknowledged or scheduled downtime.                                   |
| `--comment`          | No       | `$NOTIFICATIONCOMMENT$`            | Comment left by the author.                                                    |
| `--mention-all`      | No       | N/A                                | Boolean flag (`true`/`false`). If true, tags `@all` in the chat space.         |

---

## Nagios Configuration Examples

### 1. `commands.cfg`
Define the notification commands. Notice we create a standard command and an escalation command that includes `--mention-all=true`.

```nagios
define command {
    command_name    notify-service-by-gchat
    command_line    /usr/local/nagios/libexec/nagios-google-chat-notifications --webhook="$_CONTACTGOOGLE_CHAT_WEBHOOK$" --type="service" --nagios-url="https://your-nagios-base-url/nagios" --notification-type="$NOTIFICATIONTYPE$" --hostname="$HOSTNAME$" --hostaddress="$HOSTADDRESS$" --state="$SERVICESTATE$" --service-desc="$SERVICEDESC$" --output='$SERVICEOUTPUT$' --datetime='$LONGDATETIME$' --author='$NOTIFICATIONAUTHOR$' --comment='$NOTIFICATIONCOMMENT$'
}
define command {
    command_name    notify-service-by-gchat-escalation
    command_line    /usr/local/nagios/libexec/nagios-google-chat-notifications --webhook="$_CONTACTGOOGLE_CHAT_WEBHOOK$" --type="service" --nagios-url="https://your-nagios-base-url/nagios" --notification-type="$NOTIFICATIONTYPE$" --hostname="$HOSTNAME$" --hostaddress="$HOSTADDRESS$" --state="$SERVICESTATE$" --service-desc="$SERVICEDESC$" --output='$SERVICEOUTPUT$' --datetime='$LONGDATETIME$' --author='$NOTIFICATIONAUTHOR$' --comment='$NOTIFICATIONCOMMENT$' --mention-all=true
}

define command {
    command_name    notify-host-by-gchat
    command_line    /usr/local/nagios/libexec/nagios-google-chat-notifications --webhook="$_CONTACTGOOGLE_CHAT_WEBHOOK$" --type="host" --nagios-url="https://your-nagios-base-url/nagios" --notification-type="$NOTIFICATIONTYPE$" --hostname="$HOSTNAME$" --hostaddress="$HOSTADDRESS$" --state="$HOSTSTATE$" --output='$HOSTOUTPUT$' --datetime='$LONGDATETIME$' --author='$NOTIFICATIONAUTHOR$' --comment='$NOTIFICATIONCOMMENT$'
}
define command {
    command_name    notify-host-by-gchat-escalation
    command_line    /usr/local/nagios/libexec/nagios-google-chat-notifications --webhook="$_CONTACTGOOGLE_CHAT_WEBHOOK$" --type="host" --nagios-url="https://your-nagios-base-url/nagios" --notification-type="$NOTIFICATIONTYPE$" --hostname="$HOSTNAME$" --hostaddress="$HOSTADDRESS$" --state="$HOSTSTATE$" --output='$HOSTOUTPUT$' --datetime='$LONGDATETIME$' --author='$NOTIFICATIONAUTHOR$' --comment='$NOTIFICATIONCOMMENT$' --mention-all=true
}
```

### 2. `contacts.cfg`
Define your standard and escalation contacts. Pass the webhook URL as a custom macro.

```nagios
define contact {
    contact_name                    gchat-standard
    alias                           Google Chat Space (Standard Alerts)
    service_notification_period     24x7
    host_notification_period        24x7
    service_notification_options    w,u,c,r,f,s
    host_notification_options       d,u,r,f,s
    service_notification_commands   notify-service-by-gchat
    host_notification_commands      notify-host-by-gchat
    _GOOGLE_CHAT_WEBHOOK            <your_webhook_url>
}

define contact {
    contact_name                    gchat-escalation
    alias                           Google Chat Space (Escalation @all)
    service_notification_period     24x7
    host_notification_period        24x7
    service_notification_options    w,u,c,r
    host_notification_options       d,u,r
    service_notification_commands   notify-service-by-gchat-escalation
    host_notification_commands      notify-host-by-gchat-escalation
    _GOOGLE_CHAT_WEBHOOK            <your_webhook_url>
}
```

### 3. `escalations.cfg`
Switch to the `@all` escalation contact when a problem persists.

```nagios
define serviceescalation {
    hostgroup_name          production-servers
    service_description     *
    first_notification      3                    # Trigger on the 3rd notification
    last_notification       0                    # Continue escalating indefinitely
    notification_interval   15                   # Re-notify every 15 mins
    contacts                gchat-escalation     # Uses the command with --mention-all=true
}
```

---

## Testing CLI Command

You can easily test the binary from your terminal before hooking it into Nagios. Export your webhook URL.

```bash
export WEBHOOK="<YOUR_WEBHOOK_URL>"

./nagios-google-chat-notifications \
  --webhook="$WEBHOOK" \
  --type="service" \
  --nagios-url="https://your-nagios-base-url/nagios" \
  --notification-type="PROBLEM" \
  --hostname="my-critical-web-server" \
  --hostaddress="10.5.0.1" \
  --state="WARNING" \
  --service-desc="Disk Space" \
  --output="WARNING - 150 G (14%) free of 1069 G" \
  --datetime="Tue Jun 16 13:07:28 UTC 2026"
```
