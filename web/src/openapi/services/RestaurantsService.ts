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
     * @param authorization Bearer <Add access token here>
     * @param id Restaurant ID
     * @returns any ok
     * @throws ApiError
     */
    public static patchRestaurants(
        authorization: string,
        id: string,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/restaurants',
            headers: {
                'Authorization': authorization,
            },
            query: {
                'id': id,
            },
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
    /**
     * @param id Restaurant ID
     * @param file Menu File
     * @param token API-Token
     * @returns any ok
     * @throws ApiError
     */
    public static postRestaurants(
        id: string,
        file: Blob,
        token: string,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/restaurants/{id}',
            path: {
                'id': id,
            },
            formData: {
                'file': file,
                'token': token,
            },
            errors: {
                404: `Can not find ID`,
            },
        });
    }
}
