# sort-it

Sort-IT is a command line tool that allows you to sort your unorganized files. To do so, download the latest release binary and execute it with one of the following parameters.

```bash
# full sort, copy duplicates into subfolder of destination
--copy-duplicates
# don't check file type (mulitmedia)
--duplicates-only
# don't check file type (multimedia) and copy duplicates
--duplicates-only --copy-duplicates
# only sort files of type multimedia, ignore other file types
--multimedia-only
# only sort files of type multimedia, ignore other file types, and copy duplicates
--multimedia-only --copy-duplicates
```

When running `sort-it` it creates the following folder structure in the destination folder. Some subdirectories only get created when they are needed.

```
.
├── Data
├── Duplicates
│   ├── Files
│   ├── sort-it_duplicates.html
│   └── sort-it_duplicates.json
└── Multimedia
    ├── Audio
    │   ├── Music
    │   │   └── <Artist>
    │   │       └── <Album>
    │   └── Sounds
    │       └── <Year>
    │           └── <Month>
    ├── Pictures
    │   └── <Year>
    │       └── <Month>
    └── Videos
        └── <Year>
            └── <Month>
```

## References

- [exiftool] https://github.com/exiftool/exiftool
