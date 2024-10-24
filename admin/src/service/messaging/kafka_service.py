from typing import Iterable, Type, TypeVar

from kafka import KafkaConsumer
from loguru import logger
from py_spring_core import Component, Properties
from pydantic import BaseModel


class KafkaProperties(Properties):
    __key__: str = "Kafka"
    bootstrap_servers: str

T = TypeVar("T", bound= BaseModel)

class KafkaService(Component):
    kafka_propps: KafkaProperties
    def consume(self, topic: str, model: Type[T]) -> Iterable[T]:
        consumer = self._create_consumer()
        consumer.subscribe([topic])
        for record in consumer:
            try:
                json_value = record.value.decode("utf-8")
                yield model.model_validate_json(json_value)
            except Exception as error:
                logger.error(f"Error consuming message: {error}")
    


    def _create_consumer(self) -> KafkaConsumer:
        consumer = KafkaConsumer(bootstrap_servers=self.kafka_propps.bootstrap_servers)
        return consumer
