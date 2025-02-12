import OpenAi from 'openai';
import type { Request, Response } from 'express';

const _encodeContent = (content: string) => {
    return btoa(encodeURIComponent(content));
};

export const chat = async (req: Request, res: Response) => {
    const { message } = req.query;

    if (!message) {
        return res.end(`data: #--EOC--#\n\n`);
    }

    const openai = new OpenAi({
        apiKey: 'sk-fpcugwhxruutjioufizolnnmajwuffnnjwbtbspbilstycxp',
        baseURL: 'https://api.siliconflow.cn/v1'
    });

    res.setHeader('Content-Type', 'text/event-stream; charset=utf-8');
    res.setHeader('Cache-Control', 'no-cache');
    res.setHeader('Connection', 'keep-alive');

    const stream = await openai.chat.completions.create({
        model: 'deepseek-ai/DeepSeek-R1-Distill-Llama-8B',
        messages: [{ role: 'user', content: message as string }],
        stream: true
    });

    let isThinking = true;

    for await (const chunk of stream) {
        let data = '';
        const delta = chunk.choices[0].delta as {
            content: string;
            reasoning_content: string;
            role: string;
        };
        if (delta.reasoning_content) {
            if (!isThinking) {
                res.write(`data: #--SOT--#\n\n`);
                isThinking = true;
            }
            data = delta.reasoning_content;
        } else {
            if (isThinking) {
                res.write(`data: #--EOT--#\n\n`);
                isThinking = false;
            }
            data = delta.content;
        }
        res.write(`data: ${_encodeContent(data)}\n\n`);
    }

    res.end(`data: #--EOC--#\n\n`);
};
