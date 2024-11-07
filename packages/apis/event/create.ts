import { Event } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData
} from '@nao-todo-server/hooks';
import { ObjectId, type Oid } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const createEvent = async (req: Request, res: Response) => {
    try {
        if (!req.body.userId || !req.body.todoId || !req.body.title) {
            throw new Error('缺少参数，请求无效');
        }

        const { userId, todoId, title } = req.body;

        const createdEvent = await Event.create({
            userId,
            todoId: new ObjectId(todoId),
            title
        });

        if (!createdEvent) throw new Error('创建失败');

        return res.json(
            useSuccessfulResponseData({
                id: createdEvent._id,
                title: createdEvent.title,
                isDone: createdEvent.isDone,
                createdAt: createdEvent.createdAt,
                updatedAt: createdEvent.updatedAt
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
