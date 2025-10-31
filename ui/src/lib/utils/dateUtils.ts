/**
 * Utility functions for date and time handling
 */

/**
 * Returns a relative time display string in French based on event begin and end dates
 * Examples: "En cours", "Dans 2 heures", "Dans 30 minutes", "Terminé"
 */
export function getRelativeTimeDisplay(beginDateString: string, endDateString?: string): string {
  const now = new Date();
  const beginDate = new Date(beginDateString);
  const endDate = endDateString ? new Date(endDateString) : null;

  // Check if event is ongoing
  if (now >= beginDate && endDate && now <= endDate) {
    return "En cours";
  }

  // Event hasn't started yet
  if (now < beginDate) {
    const diffMs = beginDate.getTime() - now.getTime();
    const diffMins = Math.round(diffMs / 60000);

    if (diffMins < 60) {
      return `Dans ${diffMins} minute${diffMins !== 1 ? 's' : ''}`;
    }

    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) {
      return `Dans ${diffHours} heure${diffHours !== 1 ? 's' : ''}`;
    }

    const diffDays = Math.floor(diffHours / 24);
    return `Dans ${diffDays} jour${diffDays !== 1 ? 's' : ''}`;
  }

  // Event has ended
  return "Terminé";
}
