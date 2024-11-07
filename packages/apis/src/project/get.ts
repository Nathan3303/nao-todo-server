import { Project } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData
} from '@nao-todo-server/hooks';
import { ObjectId, Oid, serialExecute } from '@nao-todo-server/utils';
import { projectPipelines } from '@nao-todo-server/pipelines';
import type { Request, Response } from 'express';

const getProject = async (req: Request, res: Response) => {
    try {
        if (!req.query.id || !req.query.userId) {
            throw new Error('参数错误，请求无效');
        }

        const { userId, id } = req.query;

        const project = await Project.findOne({
            userId: new ObjectId(userId as string),
            _id: new ObjectId(id as string)
        });

        if (!project) throw new Error('项目不存在');

        return res.json(useSuccessfulResponseData(project));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/project/getProject] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const getProjects = async (req: Request, res: Response) => {
    try {
        const {
            userId,
            title,
            isArchived,
            isDeleted,
            isFinished,
            page,
            limit,
            sort
        } = req.query as unknown as {
            userId: Oid;
            title: string;
            isArchived: boolean;
            isDeleted: boolean;
            isFinished: boolean;
            page: string;
            limit: string;
            sort: string;
        };

        let executeResults = await serialExecute([
            () => projectPipelines.handleUserId(userId),
            () => projectPipelines.handleTitle(title),
            () => projectPipelines.handleIsDeleted(isDeleted),
            () => projectPipelines.handleIsFinished(isFinished),
            () => projectPipelines.handleIsArchived(isArchived),
            () => projectPipelines.handlePage(page, limit),
            () => projectPipelines.handleSort(sort),
            () => projectPipelines.handleOutput()
        ]);

        executeResults = executeResults.flat();
        const projects = await Project.aggregate(executeResults);

        res.status(200).json(useSuccessfulResponseData(projects));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/project/getProjects] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

export { getProject, getProjects };
