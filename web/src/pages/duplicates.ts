// Duplicate reports on the staff board (SCRUM-161). Reports in the same
// category and place whose descriptions share most of their words are one
// issue: the board shows one card, "+N more reported", and a status change or
// reassignment is done to every report in it. Routing never reads the
// description (FR-11); this happens only on the board, after routing.

// Words that say nothing about the problem itself.
const fillers = new Set([
  'the', 'and', 'for', 'are', 'was', 'were', 'been', 'has', 'have', 'had', 'not', 'its', 'this', 'that', 'there',
  'since', 'again', 'very', 'our', 'from', 'with', 'near', 'please', 'still', 'also', 'too', 'some', 'any', 'all',
  'can', 'cannot', 'into', 'out', 'any', 'one', 'now', 'today', 'yesterday',
])

// words are the meaningful words of a description: letters only, lower case,
// three letters or more, without fillers.
function words(text: string): Set<string> {
  const out = new Set<string>()
  for (const w of text.toLowerCase().split(/[^a-z]+/)) {
    if (w.length >= 3 && !fillers.has(w)) out.add(w)
  }
  return out
}

/** sameIssue: more than half of the shorter description's words are in the other. */
export function sameIssue(a: string, b: string): boolean {
  const wa = words(a)
  const wb = words(b)
  const shorter = Math.min(wa.size, wb.size)
  if (shorter === 0) return false
  let shared = 0
  for (const w of wa) if (wb.has(w)) shared++
  return shared / shorter > 0.5
}

type Place = { building?: string; wing?: string; area?: string }
type Groupable = { reportId: string; category: string; location: Place; status: string; createdAt?: string }

export type Group<T> = { leader: T; others: T[] }

const placeKey = (r: Groupable) => [r.category, r.location.building, r.location.wing, r.location.area].join('|')

/** canGroup: only reports not yet resolved are grouped. */
export const canGroup = (r: Groupable) => r.status !== 'RESOLVED'

/** samePlaceKeys gives the place of every report that shares it with another open report. */
export function sharedPlaces(reports: Groupable[]): Set<string> {
  const count = new Map<string, number>()
  for (const r of reports.filter(canGroup)) count.set(placeKey(r), (count.get(placeKey(r)) ?? 0) + 1)
  return new Set([...count].filter(([, n]) => n > 1).map(([k]) => k))
}

/** needsDescription: the board fetches the description of these reports only. */
export function needsDescription(reports: Groupable[]): string[] {
  const shared = sharedPlaces(reports)
  return reports.filter((r) => canGroup(r) && shared.has(placeKey(r))).map((r) => r.reportId)
}

/**
 * groupDuplicates puts the reports of one issue together, led by the oldest.
 * descriptions holds the description of each report that has one loaded; a
 * report without one stays on its own. Groups keep the board's order.
 */
export function groupDuplicates<T extends Groupable>(reports: T[], descriptions: Record<string, string>): Group<T>[] {
  const byAge = [...reports].sort((a, b) => (a.createdAt ?? '').localeCompare(b.createdAt ?? ''))
  const groups: Group<T>[] = []
  for (const r of byAge) {
    const text = descriptions[r.reportId]
    const home =
      canGroup(r) && text
        ? groups.find(
            (g) =>
              canGroup(g.leader) &&
              placeKey(g.leader) === placeKey(r) &&
              sameIssue(descriptions[g.leader.reportId] ?? '', text),
          )
        : undefined
    if (home) home.others.push(r)
    else groups.push({ leader: r, others: [] })
  }
  const position = new Map(reports.map((r, i) => [r.reportId, i]))
  return groups.sort((a, b) => position.get(a.leader.reportId)! - position.get(b.leader.reportId)!)
}
