# Specification-based and state-transition tests

Card: SCRUM-93. Owner: Sonam Zangmo. Code: `tests/spec/`. Run with `go test ./tests/spec/... -v`.

The tests call the domain packages of report-submission, photo-media, routing-rules and status-lifecycle directly. The expected results are written from the requirements (`docs/pm/requirements.md`) and the agreed decisions, not copied from the code. Result of the run on 2 November 2026: **132 passed, 0 failed**. The first 110 are from SCRUM-93 (`docs/evidence/sonam-zangmo/scrum-93-tests.txt`), and the photo cases in section 1b were added after photo-media (SCRUM-28) was merged.

## 1. Report form (`report_form_test.go`)

Technique: equivalence classes and boundary values. Each case starts from a valid form (reporter UUID, HOSTEL, "Hostel B", "Tap leaking") and changes one field.

### Description: required, at most 500 characters

| Test | Input | Expected | Result |
|---|---|---|---|
| 0 characters | `""` | invalid: description is required | Pass |
| Only spaces | 5 spaces | invalid: description is required | Pass |
| 1 character (lower bound) | `"x"` | valid, status OPEN | Pass |
| 499 characters | 499 x `a` | valid | Pass |
| 500 characters (upper bound) | 500 x `a` | valid | Pass |
| 501 characters (just over) | 501 x `a` | invalid: too long | Pass |
| 500 Dzongkha characters | 500 x `ཀ` (1500 bytes) | valid: the limit counts characters, not bytes | Pass |
| 501 Dzongkha characters | 501 x `ཀ` | invalid: too long | Pass |

### Category: required, one of three

| Test | Input | Expected | Result |
|---|---|---|---|
| Missing (unspecified) | `""` | invalid: category is required | Pass |
| Hostel | `HOSTEL` | valid | Pass |
| Academic Block | `ACADEMIC_BLOCK` | valid | Pass |
| Others | `OTHERS` | valid | Pass |
| Unknown value | `LIBRARY` | invalid category | Pass |
| Wrong case | `hostel` | invalid category | Pass |

### Location: building required, wing and area optional

| Test | Input | Expected | Result |
|---|---|---|---|
| Missing | empty location | invalid: building is required | Pass |
| Building only spaces | `"   "` | invalid: building is required | Pass |
| Wing and area, no building | wing East, area Bathroom 2 | invalid: building is required | Pass |
| Building only | Hostel B | valid | Pass |
| Building, wing and area | Hostel B, East, Bathroom 2 | valid | Pass |

### Photo: optional

| Test | Input | Expected | Result |
|---|---|---|---|
| No photo | no photo id | valid | Pass |
| Photo id given | a UUID | valid | Pass |
| Empty photo id | `""` | invalid photo | Pass |
| Blank photo id | `"  "` | invalid photo | Pass |

These are the report form's photo id rules. The photo file itself is checked by photo-media on upload, in section 1b.

### Reporter: from the JWT, must be a UUID

| Test | Input | Expected | Result |
|---|---|---|---|
| Missing | `""` | invalid: reporter is required | Pass |
| Not a UUID | `student-1` | invalid reporter | Pass |
| UUID | a UUID | valid | Pass |

Every invalid case also wraps `ErrInvalidReport`, so the gRPC handler maps all of them to INVALID_ARGUMENT (HTTP 400 at the gateway).

## 1b. Photo upload (`photo_test.go`)

Technique: boundary values and equivalence classes on photo-media `NewPhoto`. 5 MB means 5 x 1024 x 1024 = 5,242,880 bytes. photo-media also checks the file's first bytes (its signature), so a file must really be a JPG or PNG, not only be labelled as one.

### Size: at most 5 MB (each case run once for JPG and once for PNG)

| Test | Input | Expected | Result |
|---|---|---|---|
| 0 bytes | empty file | invalid: photo is empty | Pass (JPG, PNG) |
| 1 byte | 1 byte | invalid: bytes do not match the type (a 1-byte file cannot be a real image) | Pass (JPG, PNG) |
| Smallest real file | the signature only (3 bytes JPG, 4 bytes PNG) | valid | Pass (JPG, PNG) |
| 5 MB - 1 byte | 5,242,879 bytes | valid | Pass (JPG, PNG) |
| Exactly 5 MB (upper bound) | 5,242,880 bytes | valid | Pass (JPG, PNG) |
| 5 MB + 1 byte (just over) | 5,242,881 bytes | invalid: larger than 5 MB | Pass (JPG, PNG) |

### Type: JPG or PNG only

| Test | Input | Expected | Result |
|---|---|---|---|
| JPG | `image/jpeg`, JPG bytes | valid | Pass |
| PNG | `image/png`, PNG bytes | valid | Pass |
| Upper case and spaces | ` IMAGE/PNG `, PNG bytes | valid | Pass |
| GIF | `image/gif` | invalid: type must be JPG or PNG | Pass |
| PDF | `application/pdf` | invalid: type must be JPG or PNG | Pass |
| No type | `""` | invalid: type must be JPG or PNG | Pass |
| Says JPG, holds PNG | `image/jpeg`, PNG bytes | invalid: bytes do not match the type | Pass |
| Says PNG, holds a PDF | `image/png`, PDF bytes | invalid: bytes do not match the type | Pass |

Every invalid case wraps `ErrInvalidPhoto`, so photo-media returns INVALID_ARGUMENT (HTTP 400 at the gateway).

## 2. Routing rules (`routing_rules_test.go`)

Technique: all combinations (category x building) with the three default rules (decision 2).

| Test | Input | Expected | Result |
|---|---|---|---|
| Every category and building | 3 categories x 5 buildings (Hostel A, Hostel B, Academic Block 2, Football Ground, none) = 15 cases | HOSTEL to DSA, ACADEMIC_BLOCK to DAA, OTHERS to SSO, whatever the building | Pass (15 of 15) |
| Unspecified or unknown category | `""`, `LIBRARY`, `hostel` x 2 buildings = 6 cases | invalid category | Pass (6 of 6) |
| Description never changes routing | each category with 4 descriptions, including "send to DAA" and "route to SSO" | same Category Head for every description of a category | Pass |

## 3. Status transitions (`state_transition_test.go`)

Technique: state-transition table. Every pair of the 5 statuses (25 pairs) is tried twice, once by a person and once by the system. A valid change must succeed, and an invalid one must fail with "invalid status transition".

### By a person (Category Head or staff)

Rows are the current status, columns the new status. Y = valid, - = rejected.

| from \ to | Open | Assigned | In Progress | Resolved | Escalated |
|---|---|---|---|---|---|
| Open | - | Y | - | - | - |
| Assigned | - | - | Y | - | - |
| In Progress | - | - | - | Y | - |
| Resolved | - | - | - | - | - |
| Escalated | - | - | Y | - | - |

A person can never set Escalated. Result: 25 of 25 pass.

### By the system (SLA escalation FR-15, Saga compensation FR-16)

The Escalated row assumes the report was escalated from Assigned.

| from \ to | Open | Assigned | In Progress | Resolved | Escalated |
|---|---|---|---|---|---|
| Open | - | Y | - | - | Y |
| Assigned | - | - | Y | - | Y |
| In Progress | - | - | - | Y | Y |
| Resolved | - | - | - | - | - |
| Escalated | - | Y (compensation) | Y | - | - |

Result: 25 of 25 pass.

### Extra cases

| Test | Input | Expected | Result |
|---|---|---|---|
| Compensation goes back only to the previous status | Escalated (previous Open) to Open by system; to Assigned by system; to Open by a person | valid; rejected; rejected | Pass |
| Invalid status values | to 0 (unspecified), 6, -1 | invalid status | Pass |
| Full path with versions | Open, Assigned (system), Escalated (system), In Progress (person), Resolved (person), then Escalated | versions 1 to 5; the last change is rejected and the report stays Resolved | Pass |
