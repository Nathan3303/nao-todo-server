import mongoose from 'mongoose';
import { useErrorResponseData } from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

export const ObjectId = mongoose.Types.ObjectId;

export const checkMethod = (
    request: Request,
    response: Response,
    method: string
) => {
    if (request.method !== method) {
        response.status(405).json(useErrorResponseData('Method not allowed'));
        return false;
    }
    return true;
};

export const parseToBool = (value: unknown) => {
    if (typeof value === 'boolean') return value;
    if (typeof value === 'string') return value.trim().toLowerCase() === 'true';
    if (typeof value === 'number') return value === 1;
    return false;
};
