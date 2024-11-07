import { Tag } from '@nao-todo-server/models';
import { ObjectId } from '@nao-todo-server/utils';
import {
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

const createTag = async (req: Request, res: Response) => {
    try {
        if (!req.body.userId || !req.body.name || !req.body.color) {
            throw new Error('参数错误，请求无效');
        }

        const { userId, name, color } = req.body;

        const createdTag = await Tag.create({
            userId: new ObjectId(userId),
            name,
            color
        });
        if (!createdTag) throw new Error('创建失败');

        return res.json(
            useSuccessfulResponseData({
                id: createdTag._id,
                name: createdTag.name,
                color: createdTag.color,
                createdAt: createdTag.createdAt,
                updatedAt: createdTag.updatedAt
            })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/tag/createTag] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const createTags = async (req: Request, res: Response) => {};

export { createTag, createTags };
