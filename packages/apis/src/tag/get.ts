import { Tag } from '@nao-todo-server/models';
import { serialExecute } from '@nao-todo-server/utils';
import {
    getJWTPayload,
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { tagPipelines } from '@nao-todo-server/pipelines';
import type { Request, Response } from 'express';

const getTag = async (req: Request, res: Response) => {
    res.json(useSuccessfulResponseData('Hello, world'));
};

const getTags = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId) {
            throw new Error('参数错误，请求无效');
        }

        const { name, isDeleted, page, limit } = req.query as unknown as {
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

        const tags = await Tag.aggregate(executeResults).exec();

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
