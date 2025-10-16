# Mittagskarte

This is the open source project for the [Schniddzl.de](https://schniddzl.de) app.

## Deployment

This will pull the latest image from ghcr.io and deploy it to the provided port.
To open the app use the following link: [http://localhost:8156](http://localhost:8156)

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

## How does it work?

The app parses a config file for restaurants and their menus.
It then downloads the menus and converts them to webp files for faster loading.

For the app to understand where it has to look for the menu, it needs to be written in a special format.

To understand, please refer to the `config.go` file in the `config` directory.
There you can see a struct called Parse:

```go
type Parse struct {
	UpdateCron     string     `mapstructure:"update_cron"`
	Navigate       []Selector `mapstructure:"navigate"`
	DirectDownload string     `mapstructure:"direct_download"`
	FileType       FileType   `mapstructure:"file_type"`
}
```

The `update_cron` field defines when the app should update the menu.
Use [https://crontab.guru/](https://crontab.guru/) to generate a valid cron expression

The `direct_download` field defines where the app should look for a direct download of a menu.
The `file_type` field defines the file type of the menu ('pdf' or 'image' or leave empty for html screenscraping).

The `navigate` field defines where the app should look for the menu.
It is a list of selectors that will be used to find the menu.

Examples:

```yaml
parse:
  update_cron: "30 9,10 * * 1,3"
  direct_download: "https://davvero-stuttgart.de/download/mittagskarte.pdf"
  file_type: "pdf"
```

```yaml
parse:
  update_cron: "30 9,10 * * 1,2"
  navigate:
    - locator: ".et_pb_image_1 > span:nth-child(1) > img:nth-child(1)"
      attribute: "src"
  file_type: "image"
```

```yaml
parse:
  update_cron: "30 9,10 * * 1,4"
  navigate:
    - locator: "p.paragraph-mittagstisch-right-corona"
      style: ".w-nav { display: none !important; }"
```

```yaml
parse:
  update_cron: "30 9,10 1-3 * *"
  navigate:
    - locator: "//a[contains(text(), 'Mittagstisch') and contains(text(), '{{monthShortUpper}}')]"
  file_type: "pdf"
```

The `locator` field is an XPath or CSS selector used to locate the menu element on the page. You can find the appropriate selector using your browser’s Developer Tools (Inspect Element).

The `locator` field can contain the `date()` function, which will be replaced at runtime. It supports named arguments for flexible formatting:

| Argument | Description                                                                                                    |
| -------- | -------------------------------------------------------------------------------------------------------------- |
| `format` | Go `time` or `monday` date format string, e.g., `02.01.2006`, `Jan`, `Monday, 02 January 2006`.                |
| `lang`   | Locale/language for the output. Supported: `en`, `de`. Defaults to `en`.                                       |
| `day`    | Weekday to adjust to (`monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`, `sunday`). Optional. |
| `offset` | Number of weeks to shift the date. `-1` for last week, `0` for this week, `1` for next week. Optional.         |
| `upper`  | Set to `true` to convert the output to uppercase. Optional.                                                    |

Examples

Full month in German

```yaml
{
  # //div[@class='calendar-header']//span[text()='Oktober']
  "locator": "//div[@class='calendar-header']//span[text()='{{date(format=January, lang=de)}}']",

  # //div[@class='calendar']//span[text()='10.10.2025']
  "locator": "//div[@class='calendar']//span[text()='{{date(format=02.01.2006, day=fr, offset=-1)}}']",
}
```

The `attribute` field is an optional html attribute that will be used to find the link of the menu to be downloaded.

The `style` field is an optional style that can be used to hide elements that are not needed.

## Development

### Run locally with docker compose

When running docker compose locally it will automatically create a config.yaml file if not existing.
Changes to the file are automatically detected.

```bash
# Run dev server
docker compose up --build --force-recreate
```

### Updates

```bash
# Upgrade node packages
docker compose run --rm node yarn upgrade
# Upgrade golang packages
docker compose run --rm backend go get -u && go mod tidy
```
