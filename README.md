# goslmailer

## Drop-in notification delivery solution for slurm that can do...

* message delivery to: **msteams**, **telegram**, **e-mail**
* gathering of job **statistics**
* generating **hints** for users on how to tune their job scripts (see examples below)
* **templateable** messages ([readme](./templates/README.md))
* message **throttling**

---

## Description

**Goslmailer** (GoSlurmMailer) is a drop-in replacement [MailProg](https://slurm.schedmd.com/slurm.conf.html#OPT_MailProg) for [slurm](https://slurm.schedmd.com/).


With goslmailer configured as as the slurm mailer, 

```
MailProg                = /usr/bin/goslmailer
```

it provides users with the ability to specify a comma-separated list of receivers `[connector:]target` in the sbatch `--mail-user` switch to select where the messages will be sent out (similar to [URI scheme](https://en.wikipedia.org/wiki/Uniform_Resource_Identifier#Syntax)).

e.g.

```
sbatch --mail-type=ALL --mail-user="mailto:useremailA,msteams:usernameB,telegram:NNNNNNN,usernameC"
```

To support future additional receiver schemes, a [connector package](connectors/) has to be developed and its [configuration block](cmd/goslmailer/goslmailer.conf.annotated_example) present in configuration file.

### If you would like to contribute to this project by developing a new connector, [here](./connectors/connectorX/README.md) is a heavily annotated connector boilerplate (fully functional) to help you get started.

---

## Installation

### goslmailer

* place binary to the path of your liking
* place [goslmailer.conf](cmd/goslmailer/goslmailer.conf.annotated_example) here: `/etc/slurm/goslmailer.conf`
* point slurm `MailProg` to the binary

### gobler 

* place binary to the path of your liking
* place [gobler.conf](cmd/gobler/gobler.conf) to the path of your liking
* start the service (with -c switch pointing to config file)

### tgslurmbot

* place binary to the path of your liking
* place [tgslurmbot.conf](cmd/goslmailer/goslmailer.conf.annotated_example) to the path of your liking
  * config file has the same format as goslmailer, so you can use the same one (other connectors configs are not needed)
* start the service (with -c switch pointing to config file)

---

## Currently available connectors:

* **msteams** webhook --mail-user=`msteams:`userid
* **telegram** bot --mail-user=`telegram:`chatId
* **mailto** --mail-user=`mailto:`email-addr

See each connector details below...

## Spooling and throttling of messages - gobler service

In high-throughput clusters or in situations where job/message spikes are common, it might not be advisable to try to send all of the incoming messages as they arrive. 
For these environments goslmailer can be configured to spool messages from certain connectors on disk, to be later processed by the **gobler** service.


**gobler** is a daemon program that can be [configured](cmd/gobler/gobler.conf) to monitor specific spool directories for messages, process them and send out using the same connectors as goslmailer.


On startup, gobler reads its config file and spins-up a `connector monitor` for each configured spool directory.


`connector monitor` in turn spins up 3 goroutines: `monitor`, `picker` and `numSenders` x `sender`.

* **monitor** : 
  * every `monitorT` seconds (or milliseconds) scans the `spoolDir` for new messages and sends them to the **picker**

* **picker**  :
  * on receipt of new messages performs *trimming* of excessive messages, limiting the number of users messages in the system to `maxMsgPU`
  * every `pickerT` seconds (or milliseconds) picks the next message to be delivered and sends it to the **sender** (ordering by time of arrival)

* **sender**  :
  * `numSenders` goroutines are waiting for messages from the **picker** and try to deliver them. In case of failure, messages are returned to the **picker** for a later retry


## Artistic sketch of the system described above

![Sketch of the architecture](./images/archSketch.png)

---

## Connectors

### default connector 

Specifies which receiver scheme is the default one, in case when user didn't specify `--mail-user` and slurm sent a bare username.

```
"defaultconnector": "msteams"
```

---

### mailto

Mailto covers for original slurm e-mail sending functionality, plus a little bit more.
With connector parameters, you can:

* specify your e-mail client (ex slurm: `MailProg`, e.g. /usr/bin/mutt)
* template mail client command line  (e.g. custom subject line)
* template message body
* allowList the recipients

See [annotated configuration example](cmd/goslmailer/goslmailer.conf.annotated_example)

---

### telegram

Sends **1on1** or **group chat** messages about jobs via [telegram messenger app](https://telegram.org/)

![Telegram card](./images/telegram.png)

Prerequisites for the telegram connector: 

1. a telegram bot must be created and 
2. the bot daemon service **tgslumbot** must be running.

Site admins can [create a telegram bot](https://core.telegram.org/bots#6-botfather) by messaging [botfather](https://t.me/botfather).

Once the bot is created, you will receive a bot `token`. Place the bot `token` in the goslmailer/gobler config file in the `telegram` connector section (see example below).

Start the tgslurmbot binary that serves as the bot.

When the chat/group chat with the bot is initiated and/or the bot receives a `/start` command, he will reply with a chat-specific `--mail-user=telegram:nnn` message which the user can use in his slurm job scripts to get the job messages.

See [annotated configuration example](cmd/goslmailer/goslmailer.conf.annotated_example)

---


### msteams

Sends a message to a preconfigured ms teams channel webhook.

![MS Teams card](./images/msteams.png)

Since MS Teams does not provide with the option to send messages to users directly, only to channel webhooks, we have devised a way using MS Power Automate framework to pick up messages from this one configured *sink* channel and deliver them via private 1on1 chats to the recipient user.

Users listed in the `--mail-user=msteams:userA,msteams:userB` will be sent as adaptive card [mention](https://github.com/CLIP-HPC/goslmailer/blob/main/templates/adaptive_card_template.json#L225) entity.
A [MS Power Automate workflow](https://powerautomate.microsoft.com/en-us/) monitors the configured *sink* channel, parses the received adaptive card jsons, locates the `mention` entity and delivers to it the copy of the message via private chat.

See [annotated configuration example](cmd/goslmailer/goslmailer.conf.annotated_example)

---

## ToDo


---

## Gotchas

### msteams

* using adaptive card schema version 1.5 doesn't work with our adaptive card, check if some element changed in designer
    * tested: 1.0, 1.2 - work

## references

### msteams

* [Adaptive cards](https://adaptivecards.io/)
* [Adaptive cards - Designer](https://adaptivecards.io/designer/)
* [Rate limiting for connectors](https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/connectors-using?tabs=cURL#rate-limiting-for-connectors)
* [Bot optimizing for rate limiting](https://docs.microsoft.com/en-us/microsoftteams/platform/bots/how-to/rate-limit#)

