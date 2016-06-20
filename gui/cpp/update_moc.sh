#!/bin/bash

for f in *.h
do
	moccable=`grep Q_OBJECT "$f"`
	if [ -n "$moccable" ]
	then
		moc -qt=5 "$f" -o "moc_${f%%.h}.cpp"
		echo moc "$f" \> "moc_${f%%.h}.cpp"
	fi
done