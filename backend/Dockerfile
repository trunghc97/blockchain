# Build stage
FROM maven:3.8-openjdk-17 as build

WORKDIR /app

# Copy pom.xml for dependency resolution
COPY pom.xml .

# Download dependencies and cache them
RUN --mount=type=cache,target=/root/.m2 \
    mvn dependency:go-offline

# Copy source code
COPY src ./src

# Build application with cached dependencies
RUN --mount=type=cache,target=/root/.m2 \
    mvn clean package -DskipTests

# Production stage
FROM openjdk:17-slim
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]