import { Tag } from '@nao-todo-server/models';
import { ObjectId, Oid, serialExecute } from '@nao-todo-server/utils';
import {
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { tagPipelines } from '@nao-todo-server/pipelines';
import type { Request, Response } from 'express';

const getTag = async (req: Request, res: Response) => {};

const getTags = async (req: Request, res: Response) => {
    try {
        const { userId, name, isDeleted, page, limit } =
            req.query as unknown as {
                userId: Oid;
                name: string;
                isDeleted: boolean;
                page: string;
                limit: string;
            };

        const tasks = [
            () => tagPipelines.handleUserId(userId),
            () => tagPipelines.handleName(name),
            () => tagPipelines.handleIsDeleted(isDeleted),
            () => tagPipelines.handlePage(page, limit),
            () => tagPipelines.handleSelectFields(),
            () => Tag.aggregate().pipeline()
        ];

        let executeResults = await serialExecute(tasks);
        executeResults = executeResults.flat();
        const tags = await Tag.aggregate(executeResults);

        res.json(useSuccessfulResponseData(tags));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/tag/getTags] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

export { getTag, getTags };
