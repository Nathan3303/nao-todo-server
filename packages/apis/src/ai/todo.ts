import OpenAi from 'openai';
import {
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

export const todoChatter = async (req: Request, res: Response) => {
    const { message } = req.body;

    if (!message) {
        return res.json(useErrorResponseData('Message is required!'));
    }

    const openai = new OpenAi({
        apiKey: 'sk-fpcugwhxruutjioufizolnnmajwuffnnjwbtbspbilstycxp',
        baseURL: 'https://api.siliconflow.cn/v1'
    });

    const completion = await openai.chat.completions.create({
        model:
            process.env['OPENAI_COMPATIBLE_DEFAULT_MODEL'] ||
            'deepseek-ai/DeepSeek-R1-Distill-Llama-8B',
        messages: [{ role: 'user', content: message }]
    });

    return res.json(
        useSuccessfulResponseData(completion.choices[0].message.content)
    );
};
