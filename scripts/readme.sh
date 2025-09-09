#!/usr/bin/env bash

asciidoctorj -b docbook -o ./docs/readme.docbook ./docs/readme.adoc
echo >> ./docs/readme.docbook
pandoc -f docbook -t gfm -o ./README.md ./docs/readme.docbook --wrap=none
