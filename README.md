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
    environment:
      APP_DESCRIPTION: deine Mittagskarte für die Region Stuttgart
      APP_TITLE: Schniddzl.de
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

The `locator` field is a XPath/CSS selector that will be used to find the menu (you can use this in the developer tool of the browser to find the right XPath/CSS selector).
It can have placeholders (`{{month}}`, `{{monthShortUpper}}` or `{{year}}`) that will be replaced with for example `January`, `Jan` or `2022`.

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
