export enum DateRange {
  TODAY = 'today',
  TOMORROW = 'tomorrow',
  THIS_WEEK = 'this_week'
}


export function getMaxDateForRange(range: DateRange): Date {
  const now = new Date();
  const maxDate = new Date();
  const currentHour = now.getHours();

  switch (range) {
    case DateRange.TODAY:
      // If it's between 16h and 4h, it's "Ce soir" and we set max to 6am
      if (currentHour >= 16 || currentHour < 4) {
        if (currentHour >= 16) {
          // Between 16h and midnight - set to 6am next day
          maxDate.setDate(maxDate.getDate() + 1);
          maxDate.setHours(6, 0, 0, 0);
        } else {
          // Between midnight and 4h - set to 6am same day
          maxDate.setHours(6, 0, 0, 0);
        }
      } else {
        // Regular "Aujourd'hui" - end of today
        maxDate.setHours(23, 59, 59, 999);
      }
      break;
    case DateRange.TOMORROW:
      // End of tomorrow
      maxDate.setDate(maxDate.getDate() + 2);
      maxDate.setHours(6, 0, 0, 0);
      break;
    case DateRange.THIS_WEEK:
      // End of the week (next Sunday)
      const daysToSunday = 7 - now.getDay();
      maxDate.setDate(maxDate.getDate() + daysToSunday);
      maxDate.setHours(23, 59, 59, 999);
      break;
  }

  return maxDate;
}

/**
 * Format a time as HH:MM in French format (with 'h' separator)
 */
function formatTimeHourMinute(date: Date): string {
  const hours = date.getHours().toString().padStart(2, '0');
  const minutes = date.getMinutes().toString().padStart(2, '0');
  return `${hours}h${minutes}`;
}

/**
 * Check if a date is tomorrow
 */
function isTomorrow(date: Date): boolean {
  const now = new Date();
  const tomorrow = new Date(now);
  tomorrow.setDate(tomorrow.getDate() + 1);

  return (
    date.getDate() === tomorrow.getDate() &&
    date.getMonth() === tomorrow.getMonth() &&
    date.getFullYear() === tomorrow.getFullYear()
  );
}

function isToday(date: Date): boolean {
  const now = new Date();
  return date.getDate() === now.getDate() && date.getMonth() === now.getMonth() && date.getFullYear() === now.getFullYear();
}


function formatTimeRange(beginDate: Date, endDate: Date | null): string {
  const startTime = formatTimeHourMinute(beginDate);

  if (!endDate) {
    return startTime;
  }

  const endTime = formatTimeHourMinute(endDate);
  return `${startTime} → ${endTime}`;
}

/**
 * Returns a formatted time display string for events
 * For status: "En cours", "Terminé", or "status-upcoming"
 * For display: "14h30 → 16h00" or "Demain 14h30 → 16h00"
 */
export function getRelativeTimeDisplay(beginDateString: string, endDateString?: string): {
  status: string;
  display: string;
} {
  const now = new Date();
  const beginDate = new Date(beginDateString);
  const endDate = endDateString ? new Date(endDateString) : null;

  let timeDisplay = formatTimeRange(beginDate, endDate);

  if (isTomorrow(beginDate)) {
    timeDisplay = `Demain ${timeDisplay}`;
  } else if (!isToday(beginDate)) {
    const dayNames = ['Dimanche', 'Lundi', 'Mardi', 'Mercredi', 'Jeudi', 'Vendredi', 'Samedi'];
    const dayOfWeek = dayNames[beginDate.getDay()];
    timeDisplay = `${dayOfWeek} ${timeDisplay}`;
  }

  let status: string;

  if (now >= beginDate && endDate && now <= endDate) {
    status = "En cours";
  }
  else if (now > beginDate && (!endDate || now > endDate)) {
    status = "Terminé";
  }
  else {
    status = "status-upcoming";
  }

  return {
    status,
    display: timeDisplay
  };
}
