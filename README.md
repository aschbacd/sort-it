# sort-it

Sort your unorganized files with `sort-it` using only one command. This utility is able to find duplicates, sort multimedia files like photos, videos, and audio and also to create summary files in json as well as html where all duplicates are listed.

Get your files organized and download the binary in the [release](https://github.com/aschbacd/sort-it/releases) tab.

```
$ sort-it -h

Sort your unorganized files with sort-it using only one command. This utility
is able to find duplicates, sort multimedia files like photos, videos, and
audio and also to create summary files in json as well as html where all
duplicates are listed.

Usage:
  sort-it [source folder] [destination folder] [flags]

Flags:
      --copy-duplicates   copy duplicates to destination folder
      --duplicates-only   only look for duplicate files, do not take account of file type
  -h, --help              help for sort-it
      --multimedia-only   only sort photos, videos, and audio files, ignore all other file types
```

When running `sort-it` it creates the following folder structure in the destination folder. Some subdirectories only get created when they are needed.

```
.
├── Data
├── Errors
│   ├── Duplicates
│   ├── Files
│   ├── sort-it_duplicates.html
│   ├── sort-it_duplicates.json
│   └── sort-it_errors.json
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

## Run as Docker container

Don't want to install the binary and its dependencies on your system? No problem, just use the docker image with one of the [tags](https://hub.docker.com/r/aschbacd/sort-it/tags) available.

```bash
# set environment variables
$ SRC=<"path to source folder">
$ DST=<"path to destination folder">

# run docker container
$ docker run --rm \
  -v $SRC:$SRC \
  -v $DST:$DST \
  aschbacd/sort-it:latest $SRC $DST [flags]
```

## Dependencies

* [exiftool](https://github.com/exiftool/exiftool) (enables sort-it to get exif metadata of any file)
