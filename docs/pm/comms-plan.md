# Communications Management Plan (SCRUM-81)

Project: Jurwa, campus facility issue reporting PWA. Owner: Gayley Choden (Scrum Master). Version 1.0, 26 October 2026.

This plan says who needs what information, through which channel, how often, and how decisions and blockers are handled.

## 1. Stakeholders and what they need

| Stakeholder | What they need from us | How and when |
|---|---|---|
| Development team (4 members) | Daily task status, blockers, decisions, review requests | Group chat (daily), Jira board, GitHub pull requests, sprint events |
| Module tutors: Mr. Pema Namgay (WEB303, SWE302), Mr. Nima Yoezer (SWE303); sponsors | Evidence of progress, scope and quality; final demo | Read access to Jira and GitHub; sprint review notes in `docs/sprint-reviews/`; final presentation in Week 8 |
| Students (reporters) | A form that is quick to use; a chance to give feedback | Usability testing and pilot feedback in Week 7 |
| Category Heads: DSA, DAA, SSO | Confirm routing and the reassignment flow match how they work | Requirements interview; reassignment-flow walkthrough before the pilot |
| Hostel and academic role-holders (councillors, HoDs, estate staff) | Know what they will receive and how to sign in as staff | Consulted on the current informal process; briefed before the pilot |
| College President | Read-only overview of resolution rates and overdue trends | Sees the final presentation; no operational role |

No weekly written progress report is sent to the tutors. They follow progress through Jira, GitHub and the sprint review notes, and at assessment points.

## 2. Channels

| Channel | Used for | Rule |
|---|---|---|
| Group chat | Quick questions, stand-ups, review requests, blockers | Reply within the same day; important decisions are copied into Jira |
| Jira (project SCRUM, board 1) | Backlog, sprints, card status, hours, bugs, decisions | The source of truth for progress and for the EVM report; log hours on every card |
| GitHub (jurwaCST/Jurwa) | Code, documents, pull requests, reviews | One branch and one pull request per card, Jira key in the title |
| Calls (30 minutes) | Sprint planning, review, retrospective, team design reviews | Agenda posted in the group chat beforehand; outcome written up in the repo or Jira |
| `docs/` in the repo | Plans, designs, sprint review and retrospective notes | Every document goes through a pull request |

## 3. Meeting rhythm (5-day sprints)

Sprints are 5 days long: Sprint 1 22 to 27 October, Sprint 2 28 October to 2 November, Sprint 3 3 to 8 November.

| Event | When | Who | Output |
|---|---|---|---|
| Sprint planning | Day 1 of each sprint, 30 minutes | Whole team, led by the Scrum Master | Sprint goal and cards pulled into the sprint in Jira |
| Stand-up | Day 2 and day 4 of each sprint (twice per sprint), 15 minutes, call or written in the group chat | Whole team | Each person posts the stand-up template (section 6) |
| Sprint review | Day 5, 30 minutes | Whole team | Each person demos their cards; notes in `docs/sprint-reviews/sprint-<n>.md` |
| Retrospective | Day 5, straight after the review | Whole team | Went well, went badly, one change; actions recorded in the same file |
| Design review | When a shared piece is ready (template, contracts) | Whole team | Approval on the pull request |

## 4. Recording decisions and scope changes

- Scope is locked in the charter (risk R1). Any change to scope, architecture or a decision already made needs **sign-off from all four members**.
- To propose a change: create a Jira card labelled `decision` describing the change, the reason and the impact on schedule and effort. Each member approves in a comment on that card.
- Once approved, the Scrum Master updates the affected documents through a pull request and announces it in the group chat.
- Smaller decisions made in the group chat are copied as a comment on the related Jira card the same day, so nothing lives only in chat.

## 5. Escalation path for blockers

1. **Raise it** in the group chat as soon as you are blocked: what you need, and from whom.
2. **Owner responds** the same day (for example, the folder owner or the reviewer).
3. **Not resolved within one day:** tell the Scrum Master, who raises it at the next stand-up and moves the card or reassigns work if needed.
4. **Still blocked, or it affects the schedule or scope:** the Scrum Master and Team Lead decide, using the decision process in section 4.
5. **The team cannot resolve it:** the Team Lead contacts the relevant tutor (technical issues: Mr. Pema Namgay; project management issues: Mr. Nima Yoezer).

Never work around a blocker by editing another member's folder.

## 6. Stand-up template

Post this in the group chat on stand-up days:

```
Stand-up <date>, <name>
Yesterday: <cards worked on, what was finished>
Today: <cards I will work on>
Blockers: <none, or what I need and from whom>
```

## 7. Responsibilities

| Who | Communication responsibility |
|---|---|
| Gayley Choden (Scrum Master) | Runs sprint planning, stand-ups, review and retrospective; writes the sprint review notes; keeps the Jira board accurate; chases blockers |
| Sonam Zangmo (Team Lead) | Final say on technical disagreements within agreed scope; contact with the tutors |
| Every member | Keeps their own Jira cards and hours up to date; posts stand-ups; requests and gives reviews within a day |
