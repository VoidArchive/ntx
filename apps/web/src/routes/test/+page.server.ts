import { company } from "$lib/api/client";

export const load = async () => {
    const companies= await company.listCompanies({});
    return { companies };
};
