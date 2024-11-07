import { Event } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData
} from '@nao-todo-server/hooks';
import { ObjectId, type Oid } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const updateEvent = async (req: Request, res: Response) => {
    try {
        if (!req.body.userId || !req.body.eventId) {
            throw new Error('缺少参数，请求无效');
        }

        const { userId, eventId } = req.query;

        const updatedEvent = await Event.findOneAndUpdate(
            { userId, _id: new ObjectId(eventId as string) },
            { $set: { ...req.body, updatedAt: new Date() } }
        );

        if (!updatedEvent) throw new Error('修改失败');

        return res.json(
            useSuccessfulResponseData({
                id: updatedEvent._id,
                title: updatedEvent.title,
                isDone: updatedEvent.isDone,
                isTopped: updatedEvent.isTopped
            })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/event/updateEvent] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const updateEvents = async (req: Request, res: Response) => {};

export { updateEvent, updateEvents };
