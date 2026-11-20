import { describe, expect, it } from 'vitest'
import { groupDuplicates, sameIssue } from './duplicates'

// SCRUM-161: reports about the same problem in the same place are one issue on
// the staff board. Routing never reads the description (FR-11); this grouping
// happens after routing, only on the board.

describe('sameIssue', () => {
  it('matches descriptions that share most of their words', () => {
    expect(sameIssue('Tap leaking', 'The tap in the bathroom is leaking')).toBe(true)
    expect(sameIssue('The tap has been leaking since Monday.', 'Tap in the bathroom is leaking')).toBe(true)
    expect(sameIssue('LIGHT not working', 'the light is not working!')).toBe(true)
  })

  it('does not match different problems', () => {
    expect(sameIssue('Tap leaking', 'Window glass is broken')).toBe(false)
    expect(sameIssue('Light not working', 'Tap not working')).toBe(false)
  })

  it('never matches an empty description', () => {
    expect(sameIssue('', 'Tap leaking')).toBe(false)
    expect(sameIssue('the a is', 'the a is')).toBe(false)
  })
})

const place = { building: 'HA', wing: 'Floor 1', area: 'Left wing bathroom' }
const report = (reportId: string, createdAt: string, extra: object = {}) => ({
  reportId,
  category: 'HOSTEL',
  location: place,
  status: 'ASSIGNED',
  createdAt,
  ...extra,
})

describe('groupDuplicates', () => {
  it('puts the same issue in the same place in one group, led by the oldest report', () => {
    const reports = [
      report('b', '2026-11-21T09:00:00Z'),
      report('a', '2026-11-21T08:00:00Z'),
      report('c', '2026-11-21T10:00:00Z'),
    ]
    const groups = groupDuplicates(reports, { a: 'Tap leaking', b: 'The tap is leaking', c: 'tap leaking again' })
    expect(groups).toHaveLength(1)
    expect(groups[0].leader.reportId).toBe('a')
    expect(groups[0].others.map((r) => r.reportId)).toEqual(['b', 'c'])
  })

  it('keeps different problems in the same place apart', () => {
    const groups = groupDuplicates([report('a', '2026-11-21T08:00:00Z'), report('b', '2026-11-21T09:00:00Z')], {
      a: 'Tap leaking',
      b: 'Window glass is broken',
    })
    expect(groups.map((g) => [g.leader.reportId, g.others.length])).toEqual([
      ['a', 0],
      ['b', 0],
    ])
  })

  it('keeps the same words in another place apart', () => {
    const elsewhere = report('b', '2026-11-21T09:00:00Z', { location: { ...place, area: 'Right wing bathroom' } })
    const groups = groupDuplicates([report('a', '2026-11-21T08:00:00Z'), elsewhere], { a: 'Tap leaking', b: 'Tap leaking' })
    expect(groups).toHaveLength(2)
  })

  it('never groups a resolved report', () => {
    const done = report('b', '2026-11-21T09:00:00Z', { status: 'RESOLVED' })
    const groups = groupDuplicates([report('a', '2026-11-21T08:00:00Z'), done], { a: 'Tap leaking', b: 'Tap leaking' })
    expect(groups).toHaveLength(2)
  })

  it('keeps the order of the board, by where each group leader was', () => {
    const groups = groupDuplicates(
      [report('x', '2026-11-21T07:00:00Z', { location: { ...place, area: 'Room 101' } }), report('a', '2026-11-21T08:00:00Z')],
      { x: 'Door broken', a: 'Tap leaking' },
    )
    expect(groups.map((g) => g.leader.reportId)).toEqual(['x', 'a'])
  })
})
