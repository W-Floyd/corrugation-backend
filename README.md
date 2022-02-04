# Design

The API shall adhere to KISS - it is highly opinionated, and intentionally dumb, so as to make it easier to work on.
I will expand it only when core functionality is complete.

## What Corrugation is **NOT**

Corrugation is **NOT** a replacement to Inventree or any other parts inventory system.
It is targeted towards household items, those which need not have prices added or even require quantities specified, nor specific part numbers.
For example, I don't want to have to say how many pencils, and which brand - only that a drawer contains pencils.

## Key concepts

### Location

The core unit - most commonly a cardboard box (hence "corrugation").
Boxes may be nested (that is, a box may declare a parent box or location).
They may also be located somewhere that is not a box (for example, I might say a large box is under my bed, but "under my bed" is not itself a box)

Locations that are referenced automatically exist, and when no longer referenced no longer exist.

### Item

This is an optional inventory feature that may be a part of a location.
Items are not universal, meaning a "pencil" in one box has no bearing on a "pencil" in another box.
Items are present only for ease of records keeping, movement of items between boxes, and most importantly item searching.
Quantity is an optional attribute.

### Artifact

Artifacts may be uploaded and referenced by any location or item.
This is intended to allow linking images for any given location or item, to allow easier visualization of contents or location.
It may also allow linking documents such as user manual scans to items.