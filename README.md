## Introduction
`gem2site` is a tool to convert Gemini site(or capsule in Gemini's term) to web site.

Given a Gemini capsule like the following:
```
my_capsule
    - index.gmi
    - posts
      - post-1.gmi
    - articles
      - article-1.gmi
    - mypdf.pdf
```

`gem2site my_capsule my_site` produce contents under `my_site`:
```
my_site
    - index.html
    - posts
      - post-1.html
    - articles
      - article-1.html
    - mypdf.pdf
```

* Folder structure kept the same as source
* `.gmi` files got translated to html
* Other files are copied intact

## Status
I'm already using it for my persoal site, but there are still some work to do.

## TODOs
[ ] Customizable template
  [ ] embed a default template
  [ ] option to dump default template
  [ ] option to specify template to use
