import { Project } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData,
    getJWTPayload
} from '@nao-todo-server/hooks';
import { ObjectId, serialExecute } from '@nao-todo-server/utils';
import { projectPipelines } from '@nao-todo-server/pipelines';
import type { Request, Response } from 'express';

const getProject = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.query.projectId) {
            throw new Error('参数错误，请求无效');
        }

        const projectId = req.query.projectId as string;

        const project = await Project.findOne({
            userId,
            _id: new ObjectId(projectId)
        }).exec();

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
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId) {
            throw new Error('参数错误，请求无效');
        }

        const { title, isArchived, isDeleted, isFinished, page, limit, sort } =
            req.query as unknown as {
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
            () => projectPipelines.handleSort(sort),
            () => projectPipelines.handlePage(page, limit),
            () => projectPipelines.handleOutput()
        ]);

        executeResults = executeResults.flat();
        // console.log(executeResults);
        const projects = await Project.aggregate(executeResults).exec();

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
