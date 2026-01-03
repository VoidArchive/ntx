export function debounce<T extends (...args: Parameters<T>) => void>(
	fn: T,
	delay: number
): (...args: Parameters<T>) => void {
	let timeout: ReturnType<typeof setTimeout>;
	return (...args) => {
		clearTimeout(timeout);
		timeout = setTimeout(() => fn(...args), delay);
	};
}

export function throttle<T extends (...args: Parameters<T>) => void>(
	fn: T,
	limit: number
): (...args: Parameters<T>) => void {
	let inThrottle = false;
	return (...args) => {
		if (inThrottle) return;
		fn(...args);
		inThrottle = true;
		setTimeout(() => (inThrottle = false), limit);
	};
}
