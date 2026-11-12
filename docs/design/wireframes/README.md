# Wireframes

Low-fidelity, mobile-first wireframes for the Jurwa PWA (card SCRUM-76). They show layout and fields only, not final colours or styling. The same screens stack into wider columns on a desktop browser. Report numbers, names and counts in the images are placeholders.

Requirement ids refer to `docs/pm/requirements.md`.

## 1. Student report form (`report-form.png`)

**1a. Report an issue**

| Field | Control | Rule |
|---|---|---|
| Category | Dropdown: Hostel, Academic Block, Others | Required (FR-03) |
| Location | Dropdowns: hostel or block, wing, area | Required (FR-04) |
| Description | Multi-line text box with a live counter (`0 / 500`) | Required, at most 500 characters, whitespace-only rejected (FR-05) |
| Photo | "Add photo" button (camera or gallery) | Optional, JPG or PNG, at most 5 MB (FR-06) |

Required fields are marked with a red `*`. The Submit button stays disabled until all three required fields are filled. Only category and location are used for routing (FR-11).

**1b. Submitted.** After submit the student sees the new report id, status **OPEN**, the location and the Category Head it was routed to (FR-07, FR-12), plus the status flow.

## 2. Category Head staff queue (`staff-queue.png`)

**2a. My queue.** A list of the reports assigned to the signed-in Category Head (DAA in the example). Each card shows the report id, a status badge, category, location, age and a short description. Filter chips narrow the list by status: All, Open, Assigned, In Progress, Resolved, Escalated (FR-18). Each card has **Reassign** and **Update status** actions (FR-19, FR-20). Escalated reports have a filled badge so they stand out.

**2b. Reassign.** Picks the person to reassign to, loaded from the department directory. The DAA sees the 6 departmental HoDs (5 HoDs; Mechanical's Programme Leader acts as HoD); the DSA and SSO see their own staff lists. An optional internal note is never shown to the student (FR-21).

Staff users reached by reassignment (councillors, HoDs, estate staff) use the same queue screen, showing only reports assigned to them (FR-22, FR-23).

## 3. College President dashboard (`president-dashboard.png`)

Read-only (FR-29, FR-30), marked with a READ-ONLY label and with no buttons that change data.

- Period chips: this week, this month, all time.
- Count tiles for Open, In Progress, Resolved and Escalated.
- Table of report counts by category and status, plus the resolution rate.
- List of escalated reports with location, current holder and age.

Data comes from the analytics service.

## User flow: submit to resolved

1. **Student** signs in, fills in the report form (1a) and submits. The report is created with status **Open** (1b).
2. **System** routes it by category and location: Hostel to the DSA, Academic Block to the DAA, Others to the SSO. Status becomes **Assigned** and the SLA timer starts.
3. **Category Head** sees it in their queue (2a) and reassigns it to whoever is concerned (2b), for example the DAA to an HoD.
4. **Staff user** (or the Category Head) sets it to **In Progress** while working on it, then **Resolved** when fixed. The student is notified of each change.
5. If the SLA window passes before it is resolved, the system sets it to **Escalated** automatically. It shows with a filled badge in the queue and in the President's escalated list (3).
6. **Student** sees the final status under My reports and can rate the resolution.
