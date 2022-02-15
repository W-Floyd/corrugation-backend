# Corrugation

## Design

The API shall adhere to KISS - it is highly opinionated, and intentionally dumb, so as to make it easier to work on.
I will expand it only when core functionality is complete.

### What Corrugation is **NOT**

Corrugation is **NOT** a replacement to Inventree or any other parts inventory system.
It is targeted towards household items, those which need not have prices added or even require quantities specified, nor specific part numbers.
For example, I don't want to have to say how many pencils, and which brand - only that a drawer contains pencils.

### Key concepts

#### Entity

The core unit, may describe a location or item.
It may have a name, description, a number of artifacts, and a location.

#### Artifact

Artifacts may be uploaded and referenced by any entity.
This is intended to allow linking images for any given location or item, to allow easier visualization of contents or location.
It may also allow linking documents such as user manual scans to items.