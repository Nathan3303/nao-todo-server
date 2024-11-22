import { Event } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData,
    getJWTPayload
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

const createEvent = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.body.todoId || !req.body.title) {
            throw new Error('参数错误，请求无效');
        }

        const { todoId, title } = req.body;

        const createdEvent = await Event.create({ userId, todoId, title });

        if (!createdEvent) throw new Error('创建失败');

        return res.json(
            useSuccessfulResponseData({
                ...createdEvent.toJSON(),
                id: createdEvent._id.toString()
            })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/event/createEvent] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const createEvents = async (req: Request, res: Response) => {};

export { createEvent, createEvents };
