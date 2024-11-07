import { Project } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData
} from '@nao-todo-server/hooks';
import { ObjectId, Oid } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const updateProject = async (req: Request, res: Response) => {
    try {
        if (!req.query.projectId || req.body.title === '') {
            throw new Error('参数错误，请求无效');
        }

        const updateRes = await Project.updateOne(
            { _id: new ObjectId(req.query.projectId as string) },
            { $set: { ...req.body, updatedAt: Date.now() } }
        );

        if (!updateRes || updateRes.modifiedCount < 0)
            throw new Error('更新失败');

        return res.json(
            useSuccessfulResponseData({ projectId: req.query.projectId })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/project/updateProject] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const updateProjects = async (req: Request, res: Response) => {};

export { updateProject, updateProjects };
