# WiiNewsPR

A fork of the [NewsChannel](https://github.com/WiiLink24/NewsChannel) file generator for the Wii News Channel.

## Puerto Rico News

This fork's code is streamlined to fetch news exclusively from [El Nuevo Día RSS feeds](https://www.elnuevodia.com/rss). It generate news files for the Wii News Channel in USA/English format (because that is what my Wii is configured as), with all articles hardcoded to a San Juan, Puerto Rico geolocation.

## Usage

```bash
# Build the binary executable
go build .

# Generate news file for the current hour
./WiiNewsPR
```

Output files are saved to `./v2/1/049/news.bin.{hour}` where `1` is the code for "english" and `049` is the country code for USA.

## Debugging

Uncomment `n.debugSaveArticles()` in `sources.go` to save a JSON representations of parsed articles in the `./debug` folder.
