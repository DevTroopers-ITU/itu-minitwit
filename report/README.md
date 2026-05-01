# Report

Course report for ITU-MiniTwit. Spec: [REPORT.md](https://github.com/itu-devops/MSc_lecture_notes/blob/master/REPORT.md).

## Layout

- `report.md` — single-source for the whole report
- `images/` — figures and screenshots, referenced from `report.md`
- `build/` — output PDF goes here. Final filename: `MSc_group_q.pdf`

## Format

Markdown converted to PDF with Pandoc.

Build locally:

```
cd report
pandoc report.md --number-sections -o build/MSc_group_q.pdf
```

You need a TeX engine installed (`texlive-xetex` on Debian/Arch). CI auto-build is not wired up yet — separate PR.

## Conventions

- Word limit: 2500 (images don't count).
- Each section starts with `**Author(s):** ...` per spec (group attribution).
- Reference commits/issues by full GitHub URL or short SHA — needed for the Reflection section.
- Match figure font size to body text (see session_12 `Documentation.md`).

## Working flow

- Branch `feature/report-<topic>` per section/topic, PR into `dev`.
- Don't commit `build/MSc_group_q.pdf` until we're freezing for hand-in.

## Hand-in

Deadline: **Mon 18/5/2026 14:00**.

1. Submit PDF on WISEflow.
2. PR to `final_report_urls.py` in the lecture-notes repo with the link to our final release.
