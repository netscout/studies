# Cursor와 MCP를 활용한 프로젝트 설정

## 앉아봐라 그 시절에는 말이야...

2000년대 까지만 하더라도 꼭꼭 숨어있는 정보를 찾아내는 게 능력이었습니다. 구글에서 적절하게 검색어를 넣고 자료를 뒤지면서 남들보다 더 많은 정보를 빠르게 찾는 나름의 방법을 익혀나갔죠. 외국에서 공부하고 왔던 한 개발자에게 구글에서 "xxx"에 대해서 검색해보라고 했더니 정말 검색창에 "xxx"만 입력하는 걸 보고 경악했던 일도 있었습니다. 그 당시 구글은 대화 상대도 아니었고 그저 정보를 찾기 위한 검색툴에 불과했으니 툴을 사용해보지 않은 사람들은 사용이 어려웠던 겁니다.

그런데 ChatGPT를 기점으로 시대가 많이 변하더니 이젠 변화의 속도가 점점 가속화 되어 잠시만 눈을 떼면 따라잡기가 어렵다는 느낌이 듭니다. 심지어 이러다가 뒤쳐지는 정도가 아니라 아예 쓸모없어지는게 아닌가 하는 공포마저 들게 됩니다. 별 수 있나요... 최소한 이런 발전에 뒤쳐지지만 말자고 생각하고 노력하는 수 밖에 없죠!

## MCP가 정말 좋다며? 근데...

최근에는 MCP나 A2A 같이 에이전트의 활용에 대한 내용들이 많이 나오는 거 같습니다. A2A는 에이전트 간의 협업에 대한 거고, MCP는 에이전트가 활용할 수 있는 툴에 대한 거니까 범위는 다르지만 어쨌건 에이전트 시대에 에이전트를 좀 더 잘 활용할 수 있는 시대로 변화하고 있음은 분명합니다. 세상은 돌고 도는 거라 2000년대 초반에 SOAP를 기반으로 한 웹서비스의 느낌이 들기도 합니다만 어쨌건 구조가 비슷하게 느껴질 뿐 하는 일은 천지차이죠.

MCP에 대해 사람들이 이야기 하는 걸 들어보면 분명 엄청난 프로토콜은 맞습니다만, 저와는 상관이 없어보였습니다. 피그마에서 디자인한 내용을 MCP를 통해 프론트까지 구현한다던지 하는 건 분명 대단하지만, 저는 피그마를 사용하지 않기 때문에 "아 저게 어떤 상황에서는 미친 생산성이 나오겠구나"하는 정도였죠. 근데 진짜 누구에게나 사용가능한 미친 MCP를 발견했습니다. 아니, 이미 존재하고 있었지만 조합해보니 대단했던 거죠.

## Context7과 Sequential Thinking

[Context7](https://context7.com) 은 AI를 사용하는 데 있어 가장 중요한 부분인 context의 문제를 해결해줍니다. AI를 학습된 내용을 바탕으로 코드나 답변을 생성하기 때문에 뭔가 최신으로 업데이트된 내용은 제대로 고려하지 못합니다. Context7은 주요 프레임워크나 SDK등의 사용법을 주요 언어 및 환경 별로 어떻게 사용해야 하는지 등을 정리해서 모아놓은 사이트이고, Context7 MCP는 에이전트가 이러한 내용을 자율적으로 검색해서 구현이나 답변에 참고할 수 있도록 해줍니다. 최신 정보를 에이전트가 알아서 가져다 쓸 수 있다는 거죠!

그리고 [Sequential Thinking MCP Server](https://github.com/modelcontextprotocol/servers/tree/main/src/sequentialthinking) 는 에이전트가 사용자의 요청을 받았을 때, 구조화된 사고 과정을 거쳐서 큰 문제를 작게 쪼개고 더 나은 방법을 찾기 위해서 자신의 사고를 검증 하는 등 에이전트의 문제 해결력을 더 높여주는 MCP입니다. 이 MCP는 단독으로 사용해도 큰 효과를 주지만 Context7이랑 합치면 더 시너지가 납니다!

## 직접 해보자.

우선 [개발 동생님의 Context7 소개 영상](https://youtu.be/TrXBzzsUUY4?si=OrOAD1JO8qobF-S8) 을 보면, 두 MCP를 활용해서 아주 간단한 사칙연산을 하는 MCP 서버를 만드는 과정을 보여줍니다. 저도 한 번 따라해봤습니다. 와 정말 대박이었습니다. 이 간단한 과정을 통해서 실제 동작하는 MCP 서버를 만들 수 있다니! 이렇게 함으로써 저는 typescript를 이용해서 MCP 서버를 어떻게 만들어야 하는지 한 번에 이해할 수 있었습니다.

그리고 실제 업무나 개인적으로 스터디할 때 이 MCP를 적극활용하기 시작했습니다. 효과는 정말 대단했습니다. 그 중에 개인 스터디를 위해 준비했던 내용을 공유해보겠습니다.

## 어떤 상황이냐

최근에 mongodb에 대해서(특히 aggregation pipeline) 한 번 정리하고 싶은 마음이 생겨서 에제 코드를 작성해보고자 했습니다. 그래서 ChatGPT를 통해 기본 구조를 뽑아냈죠. 학생과 선생님 collection이 있는 구조를 설명해주고 초기 데이터를 구성하는 seed 코드를 다음과 같이 생성했습니다.

```javascript
// seed.js
// Node.js MongoDB seed script for students and teachers collections

const { MongoClient, ObjectId } = require("mongodb");
const faker = require("faker");

// Configuration
const MONGO_URI = process.env.MONGO_URI || "mongodb://localhost:27017";
const DB_NAME = process.env.DB_NAME || "school";

// Constants
const GRADES = [1, 2, 3, 4, 5, 6];
const CLASSES = ["A", "B"];

// Utility to generate random scores array
function generateScores() {
  const count = faker.datatype.number({ min: 3, max: 5 });
  return Array.from({ length: count }).map(() => ({
    math: faker.datatype.number({ min: 40, max: 100 }),
    english: faker.datatype.number({ min: 40, max: 100 }),
    coding: faker.datatype.number({ min: 40, max: 100 }),
  }));
}

async function seed() {
  const client = new MongoClient(MONGO_URI, { useUnifiedTopology: true });
  try {
    await client.connect();
    const db = client.db(DB_NAME);

    // Clear existing data
    await db.collection("students").deleteMany({});
    await db.collection("teachers").deleteMany({});

    // Generate students
    const students = [];
    GRADES.forEach((grade) => {
      CLASSES.forEach((cls) => {
        const className = `${grade}-${cls}`;
        for (let i = 0; i < 18; i++) {
          students.push({
            _id: new ObjectId(),
            name: faker.name.findName(),
            grade,
            class: className,
            scores: generateScores(),
          });
        }
      });
    });

    // Generate teachers
    const teachers = [];
    GRADES.forEach((grade) => {
      CLASSES.forEach((cls) => {
        const className = `${grade}-${cls}`;
        teachers.push({
          _id: new ObjectId(),
          name: faker.name.findName(),
          grade,
          class: className,
        });
      });
    });

    // Insert seed data
    const studentResult = await db.collection("students").insertMany(students);
    const teacherResult = await db.collection("teachers").insertMany(teachers);

    console.log(`Inserted ${studentResult.insertedCount} students`);
    console.log(`Inserted ${teacherResult.insertedCount} teachers`);
    console.log("Seeding completed successfully.");
  } catch (err) {
    console.error("Error during seeding:", err);
  } finally {
    await client.close();
  }
}

seed();
```

이 코드가 아주 정확하냐 뭐 그런건 그렇게 중요하진 않았습니다. 나중에 수정하면 되니까요. 그리고 빈 폴더를 하나 만들고 `Cursor`를 열고 다음과 같이 요청했습니다. 사용된 모델은 `Cluade 3.7 Sonnet` 입니다.

```plaintext
i need to create node.js app for studying mongodb aggregation pipeline.
so i need to create basic project structure and seed script using mongodb node.js driver.

use context7 for looking for how to use node.js mongodb driver.

use sequantial thinking mcp to plan how to create basic project structure and seeding mongodb collestion.
here's starter seed script:

// seed.js
// Node.js MongoDB seed script for students and teachers collections

const { MongoClient, ObjectId } = require('mongodb');
const faker = require('faker');

// Configuration
const MONGO_URI = process.env.MONGO_URI || 'mongodb://localhost:27017';
const DB_NAME = process.env.DB_NAME || 'school';

// Constants
const GRADES = [1, 2, 3, 4, 5, 6];
const CLASSES = ['A', 'B'];

// Utility to generate random scores array
function generateScores() {
  const count = faker.datatype.number({ min: 3, max: 5 });
  return Array.from({ length: count }).map(() => ({
    math: faker.datatype.number({ min: 40, max: 100 }),
    english: faker.datatype.number({ min: 40, max: 100 }),
    coding: faker.datatype.number({ min: 40, max: 100 })
  }));
}

async function seed() {
  const client = new MongoClient(MONGO_URI, { useUnifiedTopology: true });
  try {
    await client.connect();
    const db = client.db(DB_NAME);

    // Clear existing data
    await db.collection('students').deleteMany({});
    await db.collection('teachers').deleteMany({});

    // Generate students
    const students = [];
    GRADES.forEach(grade => {
      CLASSES.forEach(cls => {
        const className = `${grade}-${cls}`;
        for (let i = 0; i < 18; i++) {
          students.push({
            _id: new ObjectId(),
            name: faker.name.findName(),
            grade,
            class: className,
            scores: generateScores()
          });
        }
      });
    });

    // Generate teachers
    const teachers = [];
    GRADES.forEach(grade => {
      CLASSES.forEach(cls => {
        const className = `${grade}-${cls}`;
        teachers.push({
          _id: new ObjectId(),
          name: faker.name.findName(),
          grade,
          class: className
        });
      });
    });

    // Insert seed data
    const studentResult = await db.collection('students').insertMany(students);
    const teacherResult = await db.collection('teachers').insertMany(teachers);

    console.log(`Inserted ${studentResult.insertedCount} students`);
    console.log(`Inserted ${teacherResult.insertedCount} teachers`);
    console.log('Seeding completed successfully.');
  } catch (err) {
    console.error('Error during seeding:', err);
  } finally {
    await client.close();
  }
}

seed();
```

Context7 MCP를 통해 mongodb 드라이버의 사용법을 참고하고, Sequential Thinking MCP를 어떻게 작업할지 구상해보라고 했습니다. 그러자 에이전트가 MCP를 호출하기 시작합니다.

<p align="center"><img src="./01.png"></p>

mongodb와 node.js driver 연결방법, 그리고 aggregation pipeline 사용법을 참고하고 있네요!

그리고 최종적으로 다음과 같이 정리해줍니다.

```markdown
I've created a complete Node.js project for studying MongoDB aggregation pipelines. Here's what I've set up:

### Project Structure

mongodb-aggregation-study/
├── package.json # Dependencies and npm scripts
├── .gitignore # Git ignore file
├── README.md # Project documentation
└── src/
├── config/
│ └── db.js # MongoDB connection configuration
├── examples/
│ ├── basic.js # Basic aggregation examples
│ └── advanced.js # Advanced aggregation examples
├── index.js # Main application entry point
└── scripts/
└── seed.js # Database seeding script

### Key Features

1. **Database Configuration**: A reusable MongoDB connection module in `db.js`

2. **Seed Script**: Updated version of your script using `@faker-js/faker` to populate the database with students and teachers data

3. **Aggregation Examples**:

   - **Basic**: Simple examples using $match, $group, $project, and $lookup
   - **Advanced**: Complex examples with $expr, $bucket, $reduce, and $facet

4. **NPM Scripts**:
   - `npm run seed` - Populate the database
   - `npm run example:basic` - Run basic aggregation examples
   - `npm run example:advanced` - Run advanced aggregation examples
   - `npm start` - Run the main application

### Getting Started

To use this project:

1. Ensure MongoDB is running locally (or update the connection URI in your environment)
2. Install dependencies with `npm install`
3. Run `npm run seed` to populate the database
4. Try the examples with `npm run example:basic` or `npm run example:advanced`

The project provides a practical environment for learning and experimenting with MongoDB's powerful aggregation pipeline features.

mongodb-aggregation-study/
├── package.json # Dependencies and npm scripts
├── .gitignore # Git ignore file
├── README.md # Project documentation
└── src/
├── config/
│ └── db.js # MongoDB connection configuration
├── examples/
│ ├── basic.js # Basic aggregation examples
│ └── advanced.js # Advanced aggregation examples
├── index.js # Main application entry point
└── scripts/
└── seed.js # Database seeding script
```

딱히 제가 요청하지 않았던 프로젝트 구조와 기초 및 고급 예제, 그리고 커맨드 까지 생성해서 알려줍니다!

이번엔 mongodb를 도커로 실행하는 방법을 요청해보겠습니다. 이번에도 context7 MCP를 통해 사용법을 참고하라고 알려줍니다.

```plaintext
very well! i need to run mongodb using docker on local. can you add instruction in @README.md for this?

use context7 if you need to reference how to do it
```

그러자 이렇게 답변해줍니다.

```markdown
Here are clear instructions you can add to your `README.md` for running MongoDB locally using Docker. This is the most common and recommended way for local development, and it will work cross-platform.

---

## Running MongoDB Locally with Docker

If you don't have MongoDB installed on your machine, you can easily run it using Docker:

docker run --name mongodb-aggregation-study -d -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=mongo -e MONGO_INITDB_ROOT_PASSWORD=mongo mongo:latest

- This command will:
  - Pull the latest MongoDB image if you don't have it.
  - Start a MongoDB server in a container named `mongodb-aggregation-study`.
  - Expose MongoDB on the default port `27017`.
  - Set the root username and password to `mongo` (you can change these if you want).

### Connecting from your app

- By default, your `.env` should use:

  MONGO_URI=mongodb://localhost:27017
  DB_NAME=school

- If you want to use authentication, update your `.env`:

  MONGO_URI=mongodb://mongo:mongo@localhost:27017
  DB_NAME=school

  And make sure your MongoDB connection code supports authentication.

### Stopping and Removing the Container

To stop MongoDB:

docker stop mongodb-aggregation-study

To remove the container:

docker rm mongodb-aggregation-study
```

그리고 데이터 구조에 변경이 필요해서, 구조를 변경함과 동시에 예제로 같이 수정해달라고 요청했습니다.

```plaintext
currently @generateScores() generate randomly 3-5 scores for math, english, coding. but let's update this to:
{
  mid: { math, english, coding },
  final: { math, english, coding }
}

and update example aggregations if needed.
```

그러면 예제까지 싹 다 고쳐줍니다! 물론 중간 중간에 .env의 호출이 빠져서 환경 변수를 못읽어오는 바람에 mongodb에 접속이 안된다던지 하는 사소한 문제가 있긴 했지만요.

## AI 시대는 일하는 방법이 다르구나!

AI를 개발에 활용한지 2년이 다 되어가는데, 너무 많이 부족합니다. 그만큼 AI의 발전이 빠르다는 말이겠죠. 최근에는 그냥 자연어 기반으로 에이전트를 백그라운드에서 굴리는 OpenAI Codex나 Claude Code 등이 새로운 개발 트렌드로 떠오르는 거 같습니다. 이렇게 되면 에이전트를 여러 개 굴리면서 그냥 말로만 코딩하는 시대가 될 거 같기도 한데요... 아직 돈이 없어서 이런 툴들을 써보진 못했습니다. 조만간에 기회가 있으면 시도해봐야죠. 적어도 뒤쳐지지는 말아야죠!
