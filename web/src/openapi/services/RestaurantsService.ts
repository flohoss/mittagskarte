/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { services_CleanRestaurant } from '../models/services_CleanRestaurant';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class RestaurantsService {
    /**
     * Get all restaurants
     * @returns services_CleanRestaurant ok
     * @throws ApiError
     */
    public static getRestaurants(): CancelablePromise<Record<string, services_CleanRestaurant>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/restaurants',
        });
    }
    /**
     * Get a single restaurant
     * @param id Restaurant ID
     * @returns services_CleanRestaurant ok
     * @throws ApiError
     */
    public static getRestaurants1(
        id: string,
    ): CancelablePromise<services_CleanRestaurant> {
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
     * Refresh a menu
     * @param id Restaurant ID
     * @returns any ok
     * @throws ApiError
     */
    public static putRestaurants(
        id: string,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/restaurants/{id}',
            path: {
                'id': id,
            },
            errors: {
                401: `Unauthorized`,
                404: `Can not find ID`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Upload a menu
     * @param authorization Bearer <Add access token here>
     * @param id Restaurant ID
     * @param file Menu File
     * @returns services_CleanRestaurant ok
     * @throws ApiError
     */
    public static postRestaurants(
        authorization: string,
        id: string,
        file: Blob,
    ): CancelablePromise<services_CleanRestaurant> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/restaurants/{id}',
            path: {
                'id': id,
            },
            headers: {
                'Authorization': authorization,
            },
            formData: {
                'file': file,
            },
            errors: {
                401: `Unauthorized`,
                404: `Can not find ID`,
            },
        });
    }
}
