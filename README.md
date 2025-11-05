# Mittagskarte

Open-source project for [Schniddzl.de](https://schniddzl.de), fetching and displaying restaurant menus automatically.

---

## Table of Contents

1. [Deployment](#deployment)
2. [Health Check](#health-check)
3. [How It Works](#how-it-works)
4. [Configuration](#configuration)
5. [Examples](#examples)
6. [Dynamic Dates in Selectors](#dynamic-dates-in-selectors)
7. [Thumbnails](#thumbnails)
8. [Development](#development)
   - [Run Locally with Docker Compose](#run-locally-with-docker-compose)
   - [Update Dependencies](#update-dependencies)

---

## Deployment

Deploy using Docker Compose. It pulls the latest image and exposes the app on the specified port.

Open the app locally: [http://localhost:8156](http://localhost:8156)

```yaml
services:
  mittagskarte:
    image: ghcr.io/flohoss/mittagskarte:latest
    container_name: mittagskarte
    restart: always
    volumes:
      - ./config:/app/config
    ports:
      - '8156:8156'
```

---

## Health Check

**`GET /health`** or **`HEAD /health`**

Returns `200 OK` with a simple response body (`.`) to indicate the application is running.

---

## How It Works

Mittagskarte fetches restaurant menus and converts them into fast-loading **webp** images.

Key features:

- Downloads menus in PDF or image formats
- Scrapes menus from HTML pages
- Converts menus to webp, **reducing images wider than 1920px**
- Updates menus automatically based on a cron schedule

---

## Configuration

See `config/config.yaml` for a full example.

### Key Fields

| Key                | Description                                                                                           |
| ------------------ | ----------------------------------------------------------------------------------------------------- |
| `api_token`        | Required if a restaurant has **no parse section**. Used to upload a menu image manually.              |
| `impressum`        | Show legal info on the page (`enabled: true`) with `responsible` name and `email`.                    |
| `log_level`        | Logging verbosity (`debug`, `info`, `warn`, `error`)                                                  |
| `meta.title`       | Website title                                                                                         |
| `meta.description` | Suffix for the HTML description. Full HTML `<meta>` description = `{meta.title} - {meta.description}` |
| `meta.social`      | Array of social media links (optional)                                                                |
| `restaurants`      | Dictionary of restaurants                                                                             |
| `server.address`   | Host to bind (`0.0.0.0` for all interfaces)                                                           |
| `server.port`      | Port to serve the app                                                                                 |
| `time_zone`        | Used for scheduling updates                                                                           |
| `umami_analytics`  | Optional analytics integration (`enabled`, `domain`, `websiteid`)                                     |

**Notes:**

- Only `name` and `url` are strictly required for each restaurant.
- If `parse` is empty, `api_token` **must** be set.
- Menus are automatically resized to a maximum width of 1920px.

---

## Examples

**Direct PDF download**

```yaml
parse:
  update_cron: '30 9,10 * * 1,3'
  direct_download: 'https://davvero-stuttgart.de/download/mittagskarte.pdf'
  file_type: 'pdf'
```

**Image download via CSS selector**

```yaml
parse:
  update_cron: '30 9,10 * * 1,2'
  navigate:
    - locator: '.et_pb_image_1 > span:nth-child(1) > img:nth-child(1)'
      attribute: 'src'
  file_type: 'image'
```

**HTML scraping**

```yaml
parse:
  update_cron: '30 9,10 * * 1,4'
  navigate:
    - locator: 'p.paragraph-mittagstisch-right-corona'
      style: '.w-nav { display: none !important; }'
```

**PDF link via XPath**

```yaml
parse:
  update_cron: '30 9,10 1-3 * *'
  navigate:
    - locator: "//a[contains(text(), 'Mittagstisch')]"
  file_type: 'pdf'
```

All `navigate` steps use:

```yaml
- locator: '<CSS or XPath selector>'
  attribute: '<optional HTML attribute to fetch>'
  style: '<optional CSS to hide unwanted elements>'
```

---

## Dynamic Dates in Selectors

Use `{{date(...)}}` to match menus dynamically.

| Argument | Description                                                             |
| -------- | ----------------------------------------------------------------------- |
| `format` | Go time format (e.g., `02.01.2006`, `Jan`)                              |
| `lang`   | Language (`en`, `de`). Defaults to `en`                                 |
| `day`    | Weekday to adjust to (`monday`, `tuesday`, etc.)                        |
| `offset` | Number of weeks to shift (`-1` last week, `0` this week, `1` next week) |
| `upper`  | Convert output to uppercase                                             |

**Example:**

```yaml
'locator': "//div[@class='calendar']//span[text()='{{date(format=02.01.2006, day=fr, offset=-1)}}']"
```

## Thumbnails

You can add custom thumbnails for restaurants to be displayed in the app.

### How to Add Thumbnails

1. **Folder**: Place your thumbnails inside the `config/thumbnails` directory.
2. **File Name**: Name the thumbnail file exactly like the restaurant key in your `restaurants` section, with a `.webp` extension.
   - Example: For the restaurant key `sw34`, the thumbnail must be:

     ```
     config/thumbnails/sw34.webp
     ```

3. **Format**: Only **WebP** format is supported.

### Usage

The app automatically uses the thumbnail for a restaurant:

```css
background-image: url(/thumbnails/<restaurant_key>.webp);
```

Example with `sw34`:

```css
background-image: url(/thumbnails/sw34.webp);
```

> Note: WebP is required. You can convert images online using [https://mazanoke.y8o.de/](https://mazanoke.y8o.de/).

---

## Development

### Run Locally with Docker Compose

```bash
docker compose up --build --force-recreate
```

- Auto-creates `config.yaml` if missing
- Detects changes automatically

### Update Dependencies

```bash
# Node packages
docker compose run --rm node yarn upgrade --latest

# Go packages
docker compose run --rm backend go get -u && go mod tidy
```

### Build release

```sh
docker build --platform=linux/amd64 -f docker/dockerfile -t mittagskarte .
```
