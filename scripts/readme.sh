#!/usr/bin/env bash

asciidoctorj -b docbook -o README.docbook README.adoc
echo >> README.docbook
pandoc -f docbook -t gfm -o README.md README.docbook --wrap=none
