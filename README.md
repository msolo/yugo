# 🚗 yugo

`yugo` is an extremely simple static site generator.

After some hours experimenting with `hugo` and `jekyll`, `yugo` was born to solve basic publishing.

> ["Introducing the same old idea."](https://hagerty-media-prod.imgix.net/2025/04/Yugo-Ford-ModelA-VW-Bug-scaled.jpg)
>
> ["The road back to sanity."](https://hagerty-media-prod.imgix.net/2025/04/Yugo-Ford-ModelA-VW-Bug-scaled.jpg)


# Main Features
  * Simple, opinionated and useful by default.
  * Produces correct output on every build and every update.
  * Efficient live reloading during development.
  * Markdown processing for `.md` files.
    * Automatically resolves links to Markdown files so GitHub Markdown files can be published as-is.
  * Support for frontmatter in [JSONR](https://github.com/msolo/jsonr) format which gets exposed as `.Page` in templates.
  * Raw HTML passthrough in `.html` files.


# Creating A New Site

```
mkdir demo
yugo init demo
yugo serve --site demo
```

This will publish the website at http://localhost:8817/ and rebuild automatically when any changes are made.

## Publishing

The `build` mode will generate the site in `./public`  without live reloading. The `OutDir` key can be adjusted in `yugo.jsonr` to adjust where the output files are written.

```
yugo build --site demo
```


# Directory Organization

## /content

Markdown (`.md`) and HTML (`.hmtl`) files in `content` are evaluated and rendered in the `public` output directory using `templates/base.tmpl`.

Other files are copied through to `public`.

## /static

Files in `static` are copied through to the `public` output directory unmodified.

## /templates

All `.html` files here are available as templates using the Go template system.

They are automatically reloaded as they are edited.

## site.jsonr

This file sets the `.Site` variables available in all templates.

## yugo.jsonr

This file sets variables that control `yugo` itself. The presence of this file defines the root from which all other relative paths are calculated.

 - **`OutDir`** controls which directory is used for output. This is relative to the location of the site directory which contains `yugo.jsonr`.
