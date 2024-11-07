import { Project } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const createProject = async (req: Request, res: Response) => {
    try {
        if (!req.body.userId || !req.body.name) {
            throw new Error('参数错误，请求无效');
        }

        const { userId, title, description } = req.body;

        const isProjectExist = await Project.findOne({ title, userId });
        if (isProjectExist) throw new Error('项目已存在');

        const createdProject = await Project.create({
            userId: new ObjectId(userId),
            title,
            description
        });
        if (!createdProject) throw new Error('创建项目失败');

        return res.json(useSuccessfulResponseData(createdProject));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/project/createProject] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const createProjects = async (req: Request, res: Response) => {};

export { createProject, createProjects };
