import type { Period } from './settings';

export function periodToDays(period: Period): number | null {
	switch (period) {
		case '24 hours':
			return 1;
		case 'Week':
			return 7;
		case 'Month':
			return 30;
		case '3 months':
			return 90;
		case '6 months':
			return 30 * 7;
		case 'Year':
			return 365;
		default:
			return null;
	}
}

export function periodToMarkers(period: string): number | null {
	switch (period) {
		case '24h':
			return 38;
		case '7d':
			return 84;
		case '30d':
		case '60d':
			return 120;
		default:
			return null;
	}
}

export function dateInPeriod(date: Date, period: Period) {
	const days = periodToDays(period);
	if (days === null) {
		return true;
	}
	const periodAgo = new Date();
	periodAgo.setDate(periodAgo.getDate() - days);
	return date > periodAgo;
}

export function dateInPrevPeriod(date: Date, period: Period) {
	const days = periodToDays(period);
	if (days === null) {
		return true;
	}
	return dateInPrevPeriod(date, period);
}