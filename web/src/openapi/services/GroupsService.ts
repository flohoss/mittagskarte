/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { handler_Restaurant } from '../models/handler_Restaurant';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class GroupsService {
    /**
     * @returns handler_Restaurant ok
     * @throws ApiError
     */
    public static getGroups(): CancelablePromise<Record<string, Array<handler_Restaurant>>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/groups',
        });
    }
}
