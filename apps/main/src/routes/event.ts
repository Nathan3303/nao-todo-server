import express from 'express';
import {
    createEvent,
    deleteEvent,
    updateEvent,
    getEvent,
    getEvents
} from '@nao-todo-server/apis';

const exRouter: express.Router = express.Router();

exRouter.post('/event', async (req, res) => {
    await createEvent(req, res);
});

exRouter.delete('/event', async (req, res) => {
    await deleteEvent(req, res);
});

exRouter.put('/event', async (req, res) => {
    await updateEvent(req, res);
});

exRouter.get('/event', async (req, res) => {
    await getEvent(req, res);
});

exRouter.get('/events', async (req, res) => {
    await getEvents(req, res);
});

export default exRouter;
