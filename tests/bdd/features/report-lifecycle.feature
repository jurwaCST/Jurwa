Feature: Report lifecycle
  A student reports a broken campus facility. Jurwa checks the form, routes the
  report to the right Category Head, tracks its status until it is resolved and
  escalates it automatically when its SLA passes.

  Background:
    Given a student is signed in

  Rule: The report form is checked before anything is stored

    Scenario: A student submits a valid report
      When the student submits a report with category "HOSTEL", location "Hostel A / Wing 1 / Bathroom 1" and description "The tap is leaking"
      Then the report is stored
      And its status is "OPEN"

    Scenario: A report without a category is rejected
      When the student submits a report with no category, location "Hostel A" and description "The tap is leaking"
      Then the report is rejected with "category is required"
      And nothing is stored

    Scenario: A description of 501 characters is rejected
      When the student submits a report with category "HOSTEL", location "Hostel A" and a description of 501 characters
      Then the report is rejected with "description must be at most 500 characters"

    Scenario: A description of exactly 500 characters is accepted
      When the student submits a report with category "HOSTEL", location "Hostel A" and a description of 500 characters
      Then the report is stored

    Scenario: A photo over 5 MB is rejected
      When the student uploads a JPEG photo of 5242881 bytes
      Then the photo is rejected with "photo is larger than 5 MB"

    Scenario: A photo of exactly 5 MB is accepted
      When the student uploads a JPEG photo of 5242880 bytes
      Then the photo is stored
      And the student gets a photo id

  Rule: A report is routed by its category only

    Scenario Outline: A report is routed to the Category Head of its category
      When the student submits a report with category "<category>"
      Then the report is routed to "<category head>"

      Examples:
        | category       | category head |
        | HOSTEL         | DSA           |
        | ACADEMIC_BLOCK | DAA           |
        | OTHERS         | SSO           |

    Scenario: A category head reassigns a report
      Given a report with category "ACADEMIC_BLOCK" is routed to "DAA"
      When the DAA reassigns the report to the HoD of the Department of IT
      Then the report is assigned to the HoD of the Department of IT
      And the reassignment appears in the report timeline

  Rule: The status only moves forward, one step at a time

    Scenario: A report moves from Open to Resolved
      Given a report with status "OPEN"
      When the category head changes the status to "ASSIGNED"
      And the assignee changes the status to "IN_PROGRESS"
      And the assignee changes the status to "RESOLVED"
      Then the status is "RESOLVED"
      And the timeline shows "OPEN", "ASSIGNED", "IN_PROGRESS", "RESOLVED" oldest first

    Scenario: A resolved report cannot go back to Open
      Given a report with status "RESOLVED"
      When the category head changes the status to "OPEN"
      Then the change is rejected
      And the status is still "RESOLVED"

    Scenario: A step cannot be skipped
      Given a report with status "OPEN"
      When the category head changes the status to "RESOLVED"
      Then the change is rejected
      And the status is still "OPEN"

  Rule: A report whose SLA passes is escalated automatically

    Scenario Outline: A report whose SLA passes becomes Escalated
      Given a report with status "<status>"
      And its SLA window has passed
      When the SLA escalation checker runs
      Then the status is "ESCALATED"
      And no user is recorded as the actor

      Examples:
        | status      |
        | OPEN        |
        | ASSIGNED    |
        | IN_PROGRESS |

    Scenario: A resolved report is never escalated
      Given a report with status "RESOLVED"
      And its SLA window has passed
      When the SLA escalation checker runs
      Then the status is still "RESOLVED"

  Rule: Internal notes are for staff only

    Scenario: A student cannot read internal notes
      Given a staff member added the note "Plumber booked for Monday" to a report
      When the student who submitted the report asks for its notes
      Then the request is denied
