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
    if (diffHours < 12) {
      return `Dans ${diffHours} heure${diffHours !== 1 ? 's' : ''}`;
    }

    // Check if it's tomorrow
    const tomorrow = new Date(now);
    tomorrow.setDate(tomorrow.getDate() + 1);

    if (
      beginDate.getDate() === tomorrow.getDate() &&
      beginDate.getMonth() === tomorrow.getMonth() &&
      beginDate.getFullYear() === tomorrow.getFullYear()
    ) {
      // Format time as HH:MM
      const hours = beginDate.getHours().toString().padStart(2, '0');
      const minutes = beginDate.getMinutes().toString().padStart(2, '0');
      return `Demain à ${hours}:${minutes}`;
    }

    // Check if it's within the next week
    if (diffHours < 24 * 7) {
      const dayNames = ['Dimanche', 'Lundi', 'Mardi', 'Mercredi', 'Jeudi', 'Vendredi', 'Samedi'];
      const dayName = dayNames[beginDate.getDay()];
      const hours = beginDate.getHours().toString().padStart(2, '0');
      const minutes = beginDate.getMinutes().toString().padStart(2, '0');

      return `${dayName} à ${hours}:${minutes}`;
    }

    // More than a week away
    const diffDays = Math.floor(diffHours / 24);
    return `Dans ${diffDays} jour${diffDays !== 1 ? 's' : ''}`;
  }

  // Event has ended
  return "Terminé";
}
