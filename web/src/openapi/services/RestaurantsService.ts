/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { handler_Restaurant } from '../models/handler_Restaurant';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class RestaurantsService {
    /**
     * @returns handler_Restaurant ok
     * @throws ApiError
     */
    public static getRestaurants(): CancelablePromise<Record<string, handler_Restaurant>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/restaurants',
        });
    }
    /**
     * @param id Restaurant ID
     * @returns handler_Restaurant ok
     * @throws ApiError
     */
    public static getRestaurants1(
        id: string,
    ): CancelablePromise<handler_Restaurant> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/restaurants/{id}',
            path: {
                'id': id,
            },
            errors: {
                404: `Can not find ID`,
            },
        });
    }
}
