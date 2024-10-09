/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { services_DayOfWeek } from './services_DayOfWeek';
import type { services_Group } from './services_Group';
export type services_CleanRestaurant = {
    address: string;
    description: string;
    group: services_Group;
    id: string;
    image_url: string;
    name: string;
    page_url: string;
    phone: string;
    rest_days: Array<services_DayOfWeek>;
};

