export enum DateRange {
  TODAY = 'today',
  TOMORROW = 'tomorrow',
  THIS_WEEK = 'this_week'
}

export type DateWindow = {
  min: Date;
  max: Date;
};

function dayAt3am(from: Date, dayOffset: number): Date {
  const date = new Date(from);
  date.setDate(date.getDate() + dayOffset);
  date.setHours(3, 0, 0, 0);
  return date;
}

/**
 * Returns the [min, max] window during which events are displayed for the
 * given range. An event is part of the range when its [begin, end] interval
 * overlaps the window: `begin <= max && end >= min`. This keeps ongoing
 * events visible and includes night activities ending up to 3am the next day.
 */
export function getDateWindow(range: DateRange): DateWindow {
  const now = new Date();
  const min = new Date(now);

  switch (range) {
    case DateRange.TODAY: {
      // From now until 3am tomorrow (night activities included)
      return { min, max: dayAt3am(now, 1) };
    }
    case DateRange.TOMORROW: {
      // Tomorrow's day and its night, from 3am to 3am
      return { min: dayAt3am(now, 1), max: dayAt3am(now, 2) };
    }
    case DateRange.THIS_WEEK: {
      // Until Monday 3am: the next Sunday's night belongs to the week
      const daysToSunday = 7 - now.getDay();
      return { min, max: dayAt3am(now, daysToSunday + 1) };
    }
  }
}

/**
 * Formats a date the same way PocketBase stores dates
 * ("2006-01-02 15:04:05.000Z") so that filter comparisons are exact.
 */
export function formatDateForFilter(date: Date): string {
  return date.toISOString().replace('T', ' ');
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
