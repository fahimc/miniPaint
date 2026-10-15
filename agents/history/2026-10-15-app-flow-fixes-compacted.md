# Whole-image rotation and upgrade cache fixes

Image > Rotate now rotates all layer centers/orientations, swaps canvas dimensions
for quarter turns, and groups undoable actions. Existing layer rotation remains.
Unsupported nonrotatable layers receive explicit conversion guidance. Studio
account presentation uses displayName without changing authorization identity.

Initial browser acceptance exposed stale unhashed bundle caches. index.html and
Windows build script now add the bundle content digest; integrated iframe adds
app version and integrated HTML uses no-store. Mirrored FileBrowser adapters
kept aligned. Source, dist bundle and standalone EXE rebuilt locally.

Browser integration passed rotate90, undo, redo and save to original Drive file.
Saved PNG changed 320x200 to 200x320 with opaque corners. Standalone EXE built,
standalone functional UI not tested. No commit/push/release.
