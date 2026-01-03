import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { company, price } from '$lib/api/client';

export const load: PageServerLoad = async ({ params }) => {
	const { symbol } = params;

	try {
		const [companyRes, priceRes, fundamentalsRes, reportsRes] = await Promise.allSettled([
			company.getCompany({ symbol }),
			price.getPrice({ symbol }),
			company.getFundamentals({ symbol }),
			company.listReports({ symbol, limit: 20 })
		]);

		const companyData =
			companyRes.status === 'fulfilled' ? companyRes.value.company : null;

		if (!companyData) {
			throw error(404, `Company ${symbol} not found`);
		}

		return {
			company: companyData,
			price: priceRes.status === 'fulfilled' ? priceRes.value.price : null,
			fundamentals:
				fundamentalsRes.status === 'fulfilled' ? fundamentalsRes.value.fundamentals : null,
			reports: reportsRes.status === 'fulfilled' ? reportsRes.value.reports : []
		};
	} catch (err) {
		if (err && typeof err === 'object' && 'status' in err) {
			throw err;
		}
		console.error('Failed to load company data:', err);
		throw error(500, 'Failed to load company data');
	}
};
