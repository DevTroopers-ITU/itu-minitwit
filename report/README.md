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

## Schedule

Today is 2026-05-01. Hand-in is **Mon 2026-05-18 14:00**, so about 2.5 weeks.

| Date | Goal |
|------|------|
| Fri 2026-05-08 | Drafts of §1 System and §2 Process done |
| Mon 2026-05-12 | Drafts of §3 Reflection and §4 Generative AI done; CI build wired up |
| Fri 2026-05-15 | Internal review, figures finalized, word count under 2500 |
| Mon 2026-05-18 | WISEflow submission + PR to lecture-notes `final_report_urls.py` |

Dates are guidelines, not hard internal deadlines — adjust as we go.

## Hand-in

Deadline: **Mon 2026-05-18 14:00**.

1. Submit PDF on WISEflow.
2. PR to `final_report_urls.py` in the lecture-notes repo with the link to our final release.
