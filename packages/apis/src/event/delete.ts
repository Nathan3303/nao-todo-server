import { Event } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData,
    getJWTPayload
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const deleteEvent = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.query.eventId) {
            throw new Error('参数错误，请求无效');
        }

        const eventId = req.query.eventId as string;

        const deletedEvent = await Event.findOneAndDelete({
            _id: new ObjectId(eventId),
            userId
        }).exec();

        if (!deletedEvent) {
            throw new Error('删除失败');
        }

        return res.json(
            useSuccessfulResponseData({ eventId: deletedEvent._id.toString() })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/event/deleteEvent] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const deleteEvents = async (req: Request, res: Response) => {};

export { deleteEvent, deleteEvents };
