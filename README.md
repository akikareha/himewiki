# HimeWiki - A simple wiki engine with AI moderation

HimeWiki is a **simple wiki engine built with Go + PostgreSQL**.  
It features a minimal markup language called *Nomark* and optional
AI-based moderation.  

Released under the MIT License.

---

## Features

- **Lightweight** - runs as a single binary  
- **Nomark Markup** - a simple custom markup language (Markdown / Creole
  planned)  
- **PostgreSQL Storage** - stores both page content and images in the same
  DB  
- **Optional AI Filter** - integrates with OpenAI API for spam filtering
  and style unification  

---

## Requirements

- Go 1.24 or later  
- PostgreSQL 15 or later  
- Tested on **Linux** and **OpenBSD**  
- Not tested on macOS, but expected to work anywhere Go and PostgreSQL are
  available  

---

## Installation

```bash
git clone https://github.com/akikareha/himewiki.git
cd himewiki
make
```

This will build the `himewiki` binary.

---

## Database Setup

Create a PostgreSQL database named `himewiki`:  

```bash
createdb himewiki
```

Tables and indexes will be created automatically on the first run.

---

## Configuration

Copy the example config and edit it:  

```bash
cp config.yaml.example config.yaml
```

(!) `config.yaml` must exist in the **current working directory**.

### Basic Example

```yaml
app:
  mode: "devel"
  addr: ":4444"

database:
  host: "localhost"
  port: 5432
  user: "hime"
  password: "SuperStrongPassw0rd"
  name: "himewiki"
  sslmode: "disable"

site:
  base: "https://wiki.example.org/"
  name: "HimeWiki"
  card: "https://icon.example.org/hime/card.png"
```

### AI Filter (Optional)

Enable AI filtering for posts and images by setting the OpenAI API Key.  
To disable a filter, set `agent: "nil"`.

```yaml
filter:
  agent: "ChatGPT"   # use "nil" to disable
  key: "(Your OpenAI Key Here)"
  system: "You are a wiki content filter..."
  prompt: "Please rewrite in mild style..."

image-filter:
  agent: "ChatGPT"   # use "nil" to disable
  key: "(Your OpenAI Key Here)"
  max-length: 4194304
  max-size: 512
```

---

## Run

```bash
./himewiki
```

Then open your browser at `http://localhost:4444/`.

---

## License

[MIT License](LICENSE)

Author: Aki Kareha <aki@kareha.org>
