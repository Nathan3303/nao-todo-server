import { Todo } from '@nao-todo-server/models';
import { useErrorResponseData, useResponseData } from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const readTodo = async (req: Request, res: Response) => {};

const readTodos = async (req: Request, res: Response) => {};

export { readTodo, readTodos };
