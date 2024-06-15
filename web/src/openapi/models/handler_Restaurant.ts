/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { config_DayOfWeek } from './config_DayOfWeek';
import type { config_Group } from './config_Group';
import type { config_Menu } from './config_Menu';
export type handler_Restaurant = {
    address: string;
    group: config_Group;
    icon: string;
    id: string;
    menu: config_Menu;
    name: string;
    page_url: string;
    phone: string;
    rest_days: Array<config_DayOfWeek>;
};

