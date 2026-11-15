# Transcript reading (#90)

Story prose uses the owner-selected Georgia reading face, with existing Marcellus
headings and Josefin Sans controls. GM prose has paragraph spacing and a 70-character
reading measure. Player/system entries retain a distinct left rule and speaker.
Repeated badges, boxes and timestamps leave the reading flow; Turn details keeps
timestamps and mechanical receipts inspectable.

The transcript scrolls its own element, never an ancestor via scrollIntoView.
Automatic follow is immediate and only runs while the reader is within the
existing 80-pixel bottom threshold. Reading older text exposes Return to latest;
new entries mark that control without moving the reader. Durable IDs remain
attached to entries. Long-history fixtures keep all played history.

Checks include 300 persisted turns/600 entries, stable IDs after a live result,
auto-follow, reading older text, return-to-latest, and stationary submission at
phone/desktop widths. This does not implement token streaming or chapters.
