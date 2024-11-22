import { Project } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData,
    getJWTPayload
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const updateProject = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.query.projectId) {
            throw new Error('参数错误，请求无效');
        }

        if (req.body.title === '') {
            throw new Error('清单标题不能为空');
        }

        const projectId = req.query.projectId as string;

        const updatedProject = await Project.findOneAndUpdate(
            { _id: new ObjectId(projectId), userId: new ObjectId(userId) },
            { $set: { ...req.body } },
            { new: true }
        ).exec();

        if (!updatedProject) throw new Error('更新失败');

        return res.json(
            useSuccessfulResponseData({
                projectId: updatedProject._id.toString()
            })
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
