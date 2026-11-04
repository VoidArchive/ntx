import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url }) => {
	const baseUrl = url.origin;
	const apiUrl = apiBaseUrl();

	// Static pages
	const pages = ['', '/company', '/market-cap', '/mutual-funds'];

	// Fetch dynamic data
	const { company } = createApiClient(apiUrl);
	const { companies } = await company.listCompanies({ limit: 500 });

	const companyUrls = companies
		.map(
			(c) => `
		<url>
			<loc>${baseUrl}/company/${c.symbol}</loc>
			<changefreq>daily</changefreq>
			<priority>0.7</priority>
		</url>`
		)
		.join('');

	const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	${pages
		.map(
			(page) => `
		<url>
			<loc>${baseUrl}${page}</loc>
			<changefreq>daily</changefreq>
			<priority>${page === '' ? '1.0' : '0.8'}</priority>
		</url>`
		)
		.join('')}
	${companyUrls}
</urlset>`;

	return new Response(sitemap, {
		headers: {
			'Content-Type': 'application/xml',
			'Cache-Control': 'max-age=0, s-maxage=3600'
		}
	});
};
