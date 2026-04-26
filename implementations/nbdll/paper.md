# Paper Metadata

- **Title:**   Non-Blocking Doubly-Linked Lists with Good Amortized Complexity
- **Authors:** Niloufar Shafiei
- **Venue:**   OPODIS 2015
- **DOI:**     10.4230/LIPIcs.OPODIS.2015.35
- **PDF:**     https://drops.dagstuhl.de/storage/00lipics/lipics-vol046-opodis2015/LIPIcs.OPODIS.2015.35/LIPIcs.OPODIS.2015.35.pdf
- **Licence:** CC-BY 3.0

## Abstract
We present a new non-blocking doubly-linked list implementation for an asynchronous shared-memory system. It is the first such implementation for which an upper bound on amortized time complexity has been proved. In our implementation, operations access the list via cursors. Each cursor is located at an item in the list and is local to a process. In our implementation, cursors can be used to traverse and update the list, even as concurrent operations modify the list. The implementation supports two update operations, `insertBefore` and `delete`, and two move operations, `moveRight` and `moveLeft`.

## Implemented Sections
- Section 3 — Sequential Specification
- Section 4 — Non-blocking Implementation (Fig. 4)

## Known Omissions
- Memory reclamation is not explicitly handled as Go provides Garbage Collection.
