import { Event } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { serialExecute } from '@nao-todo-server/utils';
import { eventPipelines } from '@nao-todo-server/pipelines';
import type { Request, Response } from 'express';

const getEvent = async (req: Request, res: Response) => {
    res.json(useSuccessfulResponseData('Hello World'));
};

const getEvents = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId) {
            throw new Error('参数错误，请求无效');
        }

        const { todoId, title, isDone, isTopped, page, limit } =
            req.query as unknown as {
                todoId?: string;
                title?: string;
                isDone?: boolean;
                isTopped?: boolean;
                page: number;
                limit: number;
            };

        const getEventTasks = [
            () => eventPipelines.handleUserId(userId),
            () => eventPipelines.handleTodoId(todoId),
            () => eventPipelines.handleTitle(title),
            () => eventPipelines.handleIsDone(isDone),
            () => eventPipelines.handleIsTopped(isTopped),
            () => Event.aggregate().sort({ createdAt: 'asc' }).pipeline(),
            () => eventPipelines.handlePage(page, limit),
            () => eventPipelines.handleSelectFields()
        ];

        const getEventTasksExecution = await serialExecute(getEventTasks);
        const getEventPipelines = getEventTasksExecution.flat();

        const events = await Event.aggregate(getEventPipelines).exec();

        res.json(useSuccessfulResponseData(events));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/event/getEvents] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

export { getEvent, getEvents };
