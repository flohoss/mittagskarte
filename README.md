# 🍽️ Mittagskarte

Open-source project for [Schniddzl.de](https://schniddzl.de), fetching and displaying restaurant lunch menus automatically.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](go.mod)

---

## 📋 Table of Contents

1. [Deployment](#deployment)
2. [How It Works](#how-it-works)
3. [Configuration](#configuration)
   - [Key Configuration Fields](#key-configuration-fields)
   - [Restaurant Configuration](#restaurant-configuration)
   - [Parse Configuration Examples](#parse-configuration-examples)
   - [Dynamic Dates in Selectors](#dynamic-dates-in-selectors)
4. [Thumbnails](#thumbnails)
5. [Development](#development)
   - [Run Locally with Docker Compose](#run-locally-with-docker-compose)
   - [Update Dependencies](#update-dependencies)

---

## 🚀 Deployment

Deploy using Docker Compose. The application pulls the latest image and exposes the app on the specified port.

```yaml
services:
  mittagskarte:
    image: ghcr.io/flohoss/mittagskarte:latest
    container_name: mittagskarte
    restart: always
    volumes:
      - ./config:/app/config
    ports:
      - "8156:8156"
```

After starting the container, open the app at: [http://localhost:8156](http://localhost:8156)

---

## ⚙️ How It Works

Mittagskarte automatically fetches restaurant lunch menus and converts them into fast-loading **WebP** images.

### Key Features

- **Multiple Source Formats**: Downloads menus from PDFs or image files
- **HTML Scraping**: Extracts menu information directly from restaurant websites
- **Automatic Optimization**: Converts all menus to WebP format and resizes images wider than 1920px
- **Scheduled Updates**: Automatically updates menus based on configurable cron schedules
- **Browser Automation**: Uses Playwright for JavaScript-heavy websites

---

## 🔧 Configuration

Configuration is managed via `config/config.yaml`. See the included `config.yaml` for a complete working example.

### Key Configuration Fields

| Field              | Type    | Required | Description                                                                                           |
| ------------------ | ------- | -------- | ----------------------------------------------------------------------------------------------------- |
| `api_token`        | string  | ⚠️       | Required if a restaurant has **no parse section**. Used to upload menu images manually.               |
| `log_level`        | string  | No       | Logging verbosity: `debug`, `info`, `warn`, `error` (default: `info`)                                |
| `time_zone`        | string  | Yes      | Time zone for scheduling updates (e.g., `Europe/Berlin`)                                             |
| `server.address`   | string  | Yes      | Host to bind (use `0.0.0.0` for all interfaces)                                                       |
| `server.port`      | int     | Yes      | Port to serve the application                                                                         |
| `meta.title`       | string  | Yes      | Website title                                                                                         |
| `meta.description` | string  | Yes      | Description suffix for HTML meta tag. Full description: `{meta.title} - {meta.description}`          |
| `meta.social`      | array   | No       | Social media links (optional)                                                                         |
| `impressum`        | object  | No       | Legal info: `enabled` (bool), `responsible` (string), `email` (string)                               |
| `umami_analytics`  | object  | No       | Analytics integration: `enabled` (bool), `domain` (string), `website_id` (string)                    |
| `restaurants`      | map     | Yes      | Dictionary of restaurant configurations (see below)                                                   |

### Restaurant Configuration

Each restaurant requires a unique key and the following fields:

| Field         | Type     | Required | Description                                                                |
| ------------- | -------- | -------- | -------------------------------------------------------------------------- |
| `name`        | string   | Yes      | Restaurant name                                                            |
| `url`         | string   | Yes      | Restaurant website URL                                                     |
| `description` | string   | No       | Short description or cuisine type                                          |
| `address`     | string   | No       | Physical address                                                           |
| `phone`       | string   | No       | Contact phone number                                                       |
| `group`       | string   | No       | Group/category for organizing restaurants (e.g., city name)                |
| `rest_days`   | array    | No       | Days when restaurant is closed (e.g., `["Saturday", "Sunday"]`)            |
| `new`         | boolean  | No       | Mark as new restaurant                                                     |
| `parse`       | object   | No       | Configuration for automatic menu fetching (see examples below)             |

**Important Notes:**
- Only `name` and `url` are strictly required for each restaurant
- If `parse` is not provided or empty, the `api_token` **must** be set to allow manual uploads
- All menus are automatically resized to a maximum width of 1920px

### Parse Configuration Examples

The `parse` object defines how to automatically fetch and process menus. Here are the most common patterns:

#### 1. Direct PDF Download

For restaurants that provide a direct link to their menu PDF:

```yaml
parse:
  update_cron: "30 9,10 * * 1,3"
  direct_download: "https://example.com/menu.pdf"
  file_type: "pdf"
```

#### 2. Image Download via CSS Selector

For restaurants where the menu is embedded as an image on the website:

```yaml
parse:
  update_cron: "30 9,10 * * 1,2"
  navigate:
    - locator: ".menu-image > img"
      attribute: "src"
  file_type: "image"
```

#### 3. HTML Content Scraping

For restaurants that display their menu directly in HTML:

```yaml
parse:
  update_cron: "30 9,10 * * 1,4"
  navigate:
    - locator: "div.menu-content"
      style: ".header, .footer { display: none !important; }"
```

#### 4. PDF Link via XPath

For restaurants where you need to click a link or extract a PDF URL using XPath:

```yaml
parse:
  update_cron: "30 9,10 1-3 * *"
  navigate:
    - locator: "//a[contains(text(), 'Lunch Menu')]"
      attribute: "href"
  file_type: "pdf"
```

#### 5. Multi-Step Navigation with Cookie Consent

For websites requiring interaction before accessing the menu:

```yaml
parse:
  update_cron: "30 9,10 * * 1,2"
  navigate:
    - locator: "#acceptCookies"  # Click cookie consent button
    - locator: "div.weekly-menu"  # Then grab the menu
      style: ".navigation { display: none !important; }"
```

### Parse Field Reference

| Field            | Type   | Description                                                                           |
| ---------------- | ------ | ------------------------------------------------------------------------------------- |
| `update_cron`    | string | Cron expression for update schedule (e.g., `"30 9,10 * * 1-5"`)                      |
| `direct_download`| string | Direct URL to download menu file (PDF or image)                                       |
| `file_type`      | string | Type of file to process: `"pdf"` or `"image"`                                        |
| `navigate`       | array  | Sequence of browser actions to extract menu                                           |

Each item in the `navigate` array can contain:

| Field       | Type   | Description                                                                              |
| ----------- | ------ | ---------------------------------------------------------------------------------------- |
| `locator`   | string | CSS selector or XPath to find element (XPath must start with `//`)                      |
| `attribute` | string | (Optional) HTML attribute to extract from element (e.g., `"src"`, `"href"`)             |
| `style`     | string | (Optional) CSS to inject for hiding unwanted elements (e.g., navigation bars)           |

### Dynamic Dates in Selectors

Use `{{date(...)}}` placeholders to dynamically match dates in menu selectors. This is useful for websites that organize menus by date.

#### Available Parameters

| Parameter | Description                                                                  | Example Values              |
| --------- | ---------------------------------------------------------------------------- | --------------------------- |
| `format`  | Go time format string                                                        | `02.01.2006`, `Jan`, `Monday` |
| `lang`    | Language for month/day names: `en` or `de` (default: `en`)                  | `de`                        |
| `day`     | Target weekday (always lowercase)                                            | `monday`, `friday`          |
| `offset`  | Week offset: `-1` = last week, `0` = this week, `1` = next week             | `0`, `-1`, `1`              |
| `upper`   | Convert output to uppercase (boolean)                                        | `true`, `false`             |

#### Example Usage

```yaml
parse:
  navigate:
    # Match German date format for last Friday
    - locator: "//div[@class='calendar']//span[text()='{{date(format=02.01.2006, day=friday, offset=-1)}}']"
    
    # Match English month name in uppercase
    - locator: "//h2[contains(text(), '{{date(format=Jan, lang=en, upper=true)}}')]"
```

---

## 🖼️ Thumbnails

Add custom thumbnail images for restaurants to enhance the visual appearance of your menu application.

### Adding Thumbnails

1. **Location**: Place thumbnails in the `config/thumbnails/` directory
2. **Naming**: Name each file exactly like the restaurant key with `.webp` extension
   - Example: For restaurant key `sw34`, create `config/thumbnails/sw34.webp`
3. **Format**: Only **WebP** format is supported

### How It Works

The application automatically serves thumbnails at:
```
/thumbnails/<restaurant_key>.webp
```

Example CSS usage:
```css
background-image: url(/thumbnails/sw34.webp);
```

> **💡 Tip**: Convert images to WebP using online tools like [https://mazanoke.y8o.de/](https://mazanoke.y8o.de/)

---

## 🛠️ Development

### Run Locally with Docker Compose

Start the development environment with automatic reloading:

```bash
docker compose up --build --force-recreate
```

Features:
- Automatically creates `config/config.yaml` if missing
- Hot-reload for Go code changes (via Templ)
- Watches and recompiles Tailwind CSS automatically
- Backend accessible at [http://localhost:7331](http://localhost:7331)

### Update Dependencies

#### Node/JavaScript Packages

```bash
docker compose run --rm node yarn upgrade
```

#### Go Packages

```bash
docker compose run --rm backend go get -u && go mod tidy
```

---

## 📝 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📧 Support

For issues or questions, please open an issue on GitHub.
