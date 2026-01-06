import { company } from '$lib/api/client';

export const load = async () => {
	const res = await company.listCompanies({});
	return { companies: res.companies };
};
