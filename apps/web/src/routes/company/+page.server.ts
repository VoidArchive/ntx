import { company } from '$lib/api/client';

export const load = async () => {
	const res = await company.listCompanies({ limit: 500 });
	return { companies: res.companies };
};
