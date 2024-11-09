import { Event } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData,
    getJWTPayload
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const updateEvent = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.query.eventId) {
            throw new Error('参数错误，请求无效');
        }

        const eventId = req.query.eventId as string;

        const updatedEvent = await Event.findOneAndUpdate(
            { _id: new ObjectId(eventId), userId },
            { $set: { ...req.body } },
            { new: true }
        ).exec();

        if (!updatedEvent) {
            throw new Error('修改失败');
        }

        return res.json(
            useSuccessfulResponseData({ eventId: updatedEvent._id.toString() })
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
